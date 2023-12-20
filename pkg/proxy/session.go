/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"github.com/rs/zerolog"
	"net"
	"net/http"
	"syscall"
)

type session struct {
	logger zerolog.Logger
	tools  *netTools
	engine *Engine
	rule   *Rule

	scheme          string
	proxyRemoteAddr string

	clientConn   net.Conn
	clientReader *bufio.Reader

	serverRemoteAddr string
	serverHost       string
	clientTLSConfig  *tls.Config
	serverConn       net.Conn
	serverReader     *bufio.Reader

	close             bool
	callDepth         int
	postRequestAction func() error
}

// TODO: generate unique session id
func newSession(conn net.Conn, r *http.Request, e *Engine) (*session, error) {
	logger := e.logger.With().
		Str("clientAddr", conn.RemoteAddr().String()).
		Str("serverAddr", r.Host).
		Logger()

	serverHost, _, addrErr := net.SplitHostPort(r.Host)
	if addrErr != nil {
		return nil, addrErr
	}

	scheme := "https"
	if r.Method != http.MethodConnect {
		scheme = "http"
	}

	rule, hasRule := e.ruleByHost[serverHost]
	if !hasRule {
		rule = e.defaultRule
	}

	return &session{
		logger:           logger,
		clientConn:       conn,
		clientReader:     bufio.NewReader(conn),
		close:            false,
		scheme:           scheme,
		serverRemoteAddr: r.Host,
		serverHost:       serverHost,
		proxyRemoteAddr:  r.RemoteAddr,
		engine:           e,
		tools:            newNetTools(logger),
		rule:             rule,
	}, nil
}

func (s *session) clientConfigOrDefault() (*tls.Config, error) {
	if s.rule.ClientConfig != nil {
		return s.rule.ClientConfig(s.serverHost)
	}
	return s.engine.clientConfig(s.serverHost)
}

func (s *session) serverConfigOrDefault() (*tls.Config, error) {
	if s.rule.ServerConfig != nil {
		return s.rule.ServerConfig(s.serverHost)
	}
	return s.engine.serverConfig(s.serverHost)
}

func (s *session) readRequest() (*http.Request, error) {
	r, err := http.ReadRequest(s.clientReader)
	if err != nil {
		return nil, err
	}

	r.RemoteAddr = s.proxyRemoteAddr
	r.URL.Scheme = s.scheme
	r.URL.Host = s.serverRemoteAddr
	return r, nil
}

func (s *session) handle(r *http.Request) {
	c := &Context{Request: r, s: s}
	c.Next()
	if c.Response == nil {
		s.logger.Error().Msg("no-response")
		s.close = true
		c.Response = newHTTP11Response(http.StatusInternalServerError, c.Request)
	}

	defer s.tools.CloseBody(c.Response)
	c.Response.Close = s.close
	if err := c.Response.Write(s.clientConn); err != nil {
		s.close = true
		if !errors.Is(err, syscall.EPIPE) {
			withConn(s.logger.Info(), s.clientConn).Err(err).Msg("write")
		}
	}

	if s.postRequestAction != nil {
		if err := s.postRequestAction(); err != nil {
			s.close = true
		}
	}
	s.reset()
}

func (s *session) nextHandler() Handler {
	if s.callDepth < len(s.rule.Handlers) {
		h := s.rule.Handlers[s.callDepth]
		s.callDepth += 1
		return h
	}
	return s.handler
}

func (s *session) handler(c *Context) {
	resp := s.execute(c)
	c.Response = resp
}

func (s *session) reset() {
	s.callDepth = 0
	s.postRequestAction = nil
}

func (s *session) Close() {
	s.tools.CloseConn(s.clientConn)
	if s.serverConn != nil {
		s.tools.CloseConn(s.serverConn)
	}
}

func (s *session) tunnel() error {
	s.tools.Pipe(s.clientConn, s.serverConn)
	s.close = true
	return nil
}

func (s *session) clientHandshake() error {
	tlsClientConn := tls.Server(s.clientConn, s.clientTLSConfig)
	if handshakeErr := tlsClientConn.Handshake(); handshakeErr != nil {
		return handshakeErr
	}

	s.clientConn = tlsClientConn
	s.clientReader = bufio.NewReader(s.clientConn)
	return nil
}

func (s *session) execute(c *Context) *http.Response {
	if s.rule.Action == BlockAction {
		return newHTTP11Response(http.StatusForbidden, nil)
	}

	if c.Request.Method == http.MethodConnect {
		if s.rule.Action == TunnelAction {
			// TCP tunnel
			serverConn, dialErr := s.engine.dialTCP(c.Request.Host)
			if dialErr != nil {
				return s.onTCPDialError(c.Request, dialErr)
			}
			s.serverConn = serverConn
			s.postRequestAction = s.tunnel
			return newHTTP10ConnectionEstablished(c.Request)
		}

		serverConfig, serverConfigErr := s.serverConfigOrDefault()
		if serverConfigErr != nil {
			return s.onTLSConfigError(c.Request, serverConfigErr)
		}

		serverConn, dialErr := s.engine.dialTLS(c.Request.Host, serverConfig)
		if dialErr != nil {
			var headerErr tls.RecordHeaderError
			if errors.As(dialErr, &headerErr) {
				withConn(s.logger.Info(), headerErr.Conn).
					Bytes("recordHeader", headerErr.RecordHeader[:]).
					Str("msg", headerErr.Msg).
					Msg("tls-not-supported")

				tcpServerConn, tcpDialErr := s.engine.dialTCP(c.Request.Host)
				if tcpDialErr != nil {
					return s.onTCPDialError(c.Request, tcpDialErr)
				}

				clientConfig, clientConfigErr := s.clientConfigOrDefault()
				if clientConfigErr != nil {
					return s.onTLSConfigError(c.Request, clientConfigErr)
				}

				s.serverConn = tcpServerConn
				s.serverReader = bufio.NewReader(s.serverConn)
				s.clientTLSConfig = clientConfig
				s.postRequestAction = s.clientHandshake
				return newHTTP10ConnectionEstablished(c.Request)
			}

			var verificationErr *tls.CertificateVerificationError
			if errors.As(dialErr, &verificationErr) {
				return s.onCertificateVerificationFailure(c.Request, verificationErr)
			}

			return s.onTCPDialError(c.Request, dialErr)
		}

		s.serverConn = serverConn
		s.serverReader = bufio.NewReader(s.serverConn)
		clientConfig, clientConfigErr := s.clientConfigOrDefault()
		if clientConfigErr != nil {
			return s.onTLSConfigError(c.Request, clientConfigErr)
		}
		s.clientTLSConfig = clientConfig
		s.postRequestAction = s.clientHandshake
		return newHTTP10ConnectionEstablished(c.Request)
	}

	// http method is other than CONNECT
	if s.serverConn == nil {
		serverConn, dialErr := s.engine.dialTCP(c.Request.Host)
		if dialErr != nil {
			return s.onTCPDialError(c.Request, dialErr)
		}
		s.serverConn = serverConn
		s.serverReader = bufio.NewReader(serverConn)
	}

	writeErr := c.Request.Write(s.serverConn)
	if writeErr != nil {
		return s.onWriteErr(c.Request, writeErr)
	}

	resp, readErr := http.ReadResponse(s.serverReader, c.Request) // TODO: handle read timeout
	if readErr != nil {
		return s.onReadError(c.Request, readErr)
	}

	if isWebsocketUpgrade(c.Request) {
		s.postRequestAction = s.tunnel
	}

	return resp
}

func (s *session) onWriteErr(r *http.Request, e error) *http.Response {
	s.close = true
	withConn(s.logger.Info(), s.serverConn).Err(e).Msg("write")
	return newHTTP11Response(http.StatusBadGateway, r)
}

func (s *session) onReadError(r *http.Request, e error) *http.Response {
	s.close = true
	withConn(s.logger.Info(), s.serverConn).Err(e).Msg("read")
	return newHTTP11Response(http.StatusBadGateway, r)
}

func (s *session) onCertificateVerificationFailure(r *http.Request, e *tls.CertificateVerificationError) *http.Response {
	s.close = true

	event := s.logger.Error()
	if len(e.UnverifiedCertificates) > 0 {
		cert := e.UnverifiedCertificates[0]
		event = event.Strs("dnsNames", cert.DNSNames)
		for _, ip := range cert.IPAddresses {
			event = event.IPAddr("ip", ip)
		}
	}
	event.Msg("certificate-verification")

	return newHTTP11Response(http.StatusBadGateway, r)
}

func (s *session) onTLSConfigError(r *http.Request, e error) *http.Response {
	s.close = true
	s.logger.Error().Err(e).Str("host", r.Host).Msg("get-tls-config")
	return newHTTP11Response(http.StatusInternalServerError, r)
}

func (s *session) onTCPDialError(r *http.Request, e error) *http.Response {
	s.close = true

	event := s.logger.Info()
	var networkErr *net.OpError
	if errors.As(e, &networkErr) {
		event = withOpError(event, networkErr)
	} else {
		event = event.Err(e)
	}
	event.Msg("dial")

	var statusCode int
	if errors.Is(e, context.DeadlineExceeded) {
		statusCode = http.StatusGatewayTimeout
	} else {
		statusCode = http.StatusBadGateway
	}
	return newHTTP11Response(statusCode, r)
}

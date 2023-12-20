/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy

import (
	"crypto/elliptic"
	"crypto/tls"
	"errors"
	"github.com/pmateusz/glove/internal/ca"
	"github.com/pmateusz/glove/internal/logging"
	"github.com/pmateusz/glove/internal/runtime"
	"github.com/rs/zerolog"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"
)

// TODO: improve error handling for closed connections and writes to closed connections

type Engine struct {
	logger zerolog.Logger
	dialer *net.Dialer
	tools  *netTools

	clientConfig func(host string) (*tls.Config, error)
	serverConfig func(host string) (*tls.Config, error)

	defaultRule *Rule
	ruleByHost  map[string]*Rule
}

func (e *Engine) dialTCP(host string) (net.Conn, error) {
	return e.dialer.Dial("tcp", host)
}

func (e *Engine) dialTLS(host string, config *tls.Config) (net.Conn, error) {
	return tls.DialWithDialer(e.dialer, "tcp", host, config)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientConn, hijackErr := e.tools.Hijack(w)
	if hijackErr != nil {
		e.logger.Error().Err(hijackErr).Msg("hijack-connection")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s, createSessionErr := newSession(clientConn, r, e)
	if createSessionErr != nil {
		var addErr *net.AddrError
		if errors.As(createSessionErr, &addErr) {
			withRemoteAddr(e.logger.Info(), addErr.Addr).Err(addErr).Msg("parse-host")
			e.tools.WriteHTTP11Status(clientConn, http.StatusBadRequest)
		}
		e.tools.CloseConn(clientConn)
		return
	}

	defer s.Close()

	s.handle(r)
	for !s.close {
		req, readErr := s.readRequest() // TODO: add timeout as it hangs if request doesn't end with /r/n/r/n
		if readErr != nil {
			if errors.Is(readErr, io.EOF) || errors.Is(readErr, syscall.ECONNRESET) {
				break
			}

			withConn(s.logger.Info(), s.clientConn).Err(readErr).Msg("read")
			s.close = true
			s.tools.WriteHTTP11Status(s.clientConn, http.StatusBadRequest)
			break
		}

		s.close = req.Close
		s.handle(req)
	}
}

func NewEngine(opts ...EngineOption) *Engine {
	options := NewEngineOptions()

	for _, opt := range opts {
		opt(options)
	}

	var logger zerolog.Logger
	if options.logger == nil {
		isContainer, _ := runtime.IsContainer()
		if isContainer {
			logger = logging.NewStructLogger()
		} else {
			logger = logging.NewConsoleLogger()
		}
	} else {
		logger = *options.logger
	}

	if options.clientConfig == nil {
		defaultCA, caErr := newSelfSignedCA()
		if caErr != nil {
			options.logger.Fatal().Err(caErr).Msg("create self-signed CA")
		}

		options.clientConfig = func(host string) (*tls.Config, error) {
			cert, err := defaultCA.SignHosts(host)
			if err != nil {
				return nil, err
			}
			return &tls.Config{Certificates: []tls.Certificate{*cert}}, nil
		}
	}

	if options.serverConfig == nil {
		options.serverConfig = func(host string) (*tls.Config, error) {
			return &tls.Config{}, nil
		}
	}

	if options.dialer == nil {
		options.dialer = &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
	}

	if options.defaultRule == nil {
		options.defaultRule = &Rule{
			Action: TunnelAction,
		}
	}

	return &Engine{
		logger:       logger,
		dialer:       options.dialer,
		tools:        newNetTools(logger),
		clientConfig: options.clientConfig,
		serverConfig: options.serverConfig,
		defaultRule:  options.defaultRule,
		ruleByHost:   options.ruleByHost,
	}
}

func newSelfSignedCA() (*ca.CA, error) {
	generator := ca.ECDSAKeyGenerator{Curve: elliptic.P256()}
	return ca.NewCA(&generator, nil)
}

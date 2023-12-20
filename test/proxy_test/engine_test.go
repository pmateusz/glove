/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"github.com/pmateusz/glove/pkg/proxy"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Test Matrix:
// HTTP  proxy forward        -> HTTP  server
// HTTPS proxy forward        -> HTTP  server
// HTTP  proxy tunnel         -> HTTP  server - not implemented by the engine
// HTTPS proxy tunnel         -> HTTP  server - not implemented by the engine
// HTTP  proxy tunnel         -> HTTPS server
// HTTPS proxy tunnel         -> HTTPS server
// HTTP  proxy forward (mitm) -> HTTPS server
// HTTPS proxy forward (mitm) -> HTTPS server
// HTTP  proxy forward        -> WS    server - not working ???
// HTTPS proxy forward        -> WS    server - not supported by the websocket client
// HTTP  proxy tunnel         -> WS    server
// HTTPS proxy tunnel         -> WS    server - not supported by the websocket client
// HTTP  proxy tunnel         -> WSS   server
// HTTPS proxy tunnel         -> WSS   server - not supported by the websocket client
// HTTP  proxy forward (mitm) -> WSS   server
// HTTPS proxy forward (mitm) -> WSS   server - not supported by the websocket client

// HTTP forward proxy -> HTTP server
func TestPlainEngineHTTPProxyToHTTP(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine())
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN-THEN
	tools.AssertHTTPEcho(server.URL, "http proxy to http")
}

// HTTPS forward proxy -> HTTP server
func TestPlainEngineHTTPSProxyToHTTP(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewTLSServer(proxy.NewEngine())
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, proxyServer)

	// WHEN-THEN
	tools.AssertHTTPEcho(server.URL, "https proxy to http")
}

// HTTP forward proxy -> WS server
func TestPlainEngineHTTPProxyToWS(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine())
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN-THEN
	tools.AssertWebsocketEcho(server.URL, "http proxy to ws")
}

// HTTP forward proxy tunnel -> WS server
func TestPlainEngineHTTPProxyTunnelToWS(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine())
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN-THEN
	tools.AssertWebsocketEcho(server.URL, "http proxy tunnel to ws")
}

// HTTPS forward proxy -> WS server - not supported by the websocket client
func TestPlainEngineHTTPSProxyToWS(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewTLSServer(proxy.NewEngine())
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, proxyServer)

	// WHEN
	conn, resp, err := tools.DialWebsocket(server.URL)

	// THEN
	assert.Nil(t, conn)
	assert.Nil(t, resp)
	assert.Errorf(t, err, "proxy: unknown scheme: https")
}

// HTTP forward proxy tunnel -> HTTPS server
func TestPlainEngineHTTPProxyTunnelToHTTPS(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine())
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, server)

	// WHEN-THEN
	tools.AssertHTTPEcho(server.URL, "http proxy tunnel to https")
}

// HTTPS forward proxy tunnel -> HTTPS server
func TestPlainEngineHTTPSProxyTunnelToHTTPS(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewTLSServer(proxy.NewEngine())
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, proxyServer, server)

	// WHEN-THEN
	tools.AssertHTTPEcho(server.URL, "https proxy tunnel to https")
}

// HTTP forward proxy tunnel -> WSS server
func TestPlainEngineHTTPProxyTunnelToWSS(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine())
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, server)

	// WHEN-THEN
	tools.AssertHTTPEcho(server.URL, "http proxy tunnel to wss")
}

// HTTPS forward proxy tunnel -> WSS server - not supported by the websocket client
func TestPlainEngineHTTPSProxyTunnelToWSSNotSupportedByClient(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewTLSServer(proxy.NewEngine())
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, proxyServer, server)

	// WHEN
	conn, resp, err := tools.DialWebsocket(server.URL)

	// THEN
	assert.Nil(t, conn)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "proxy: unknown scheme: https")
}

// HTTP forward proxy mitm -> HTTPS server
func TestPlainEngineHTTPProxyMITMToHTTPS(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine(WithTestServer(server, true)))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, server)

	// WHEN-THEN
	tools.AssertHTTPEcho(server.URL, "http proxy mitm to https")
}

// HTTPS forward proxy mitm -> HTTPS server
func TestPlainEngineHTTPSProxyMITMToHTTPS(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewTLSServer(proxy.NewEngine(WithTestServer(server, true)))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, proxyServer, server)

	// WHEN-THEN
	tools.AssertHTTPEcho(server.URL, "https proxy mitm to https")
}

// HTTP forward proxy mitm -> WS server
func TestPlainEngineHTTPProxyMITMToWS(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine(WithTestServer(server, true), proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN
	_, _, err := tools.DialWebsocket(server.URL)

	// THEN
	// configuration with MITM suggests we will want to use TLS whereas the client doesn't expect it
	assert.Error(t, err, "unexpected EOF")
}

// HTTP forward proxy mitm -> WSS server
func TestPlainEngineHTTPProxyMITMToWSS(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine(WithTestServer(server, true)))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, server)

	// WHEN-THEN
	tools.AssertHTTPEcho(server.URL, "http proxy mitm to wss")
}

// HTTPS forward proxy mitm -> WSS server - not supported by the websocket client
func TestPlainEngineHTTPSProxyMITMToWSSNotSupportedByClient(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewTLSServer(proxy.NewEngine(WithTestServer(server, true)))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, proxyServer, server)

	// WHEN
	conn, resp, err := tools.DialWebsocket(server.URL)

	// THEN
	assert.Nil(t, conn)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "proxy: unknown scheme: https")
}

func TestPlainEngineHTTPProxyMITMToBadGateway(t *testing.T) {
	// GIVEN
	proxyServer := httptest.NewServer(proxy.NewEngine(WithMITM(localhost), proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN
	_, err := tools.HTTPEcho("https://127.0.0.1:9999", "http proxy mitm to bad gateway")

	// THEN
	assert.Error(t, err, "Bad Request")
}

func TestPlainEngineHTTPProxyTunnelToBadGateway(t *testing.T) {
	// GIVEN
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN
	_, err := tools.HTTPEcho("https://127.0.0.1:9999", "http proxy tunnel to bad gateway")

	// THEN
	assert.Error(t, err, "Bad Request")
}

func TestPlainEngineHTTPProxyToBadGateway(t *testing.T) {
	// GIVEN
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN
	resp, err := tools.HTTPEcho("http://127.0.0.1:9999", "http proxy to bad gateway")

	// THEN
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, resp.StatusCode, http.StatusBadGateway)
	}
}

func TestPlainEngineHTTPProxyTunnelToUnknownAuthority(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine())
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN
	_, err := tools.HTTPEcho(server.URL, "http proxy tunnel to HTTPS with unknown authority")

	// THEN
	var tlsErr *tls.CertificateVerificationError
	if assert.ErrorAs(t, err, &tlsErr) {
		var x509Err x509.UnknownAuthorityError
		assert.ErrorAs(t, tlsErr.Err, &x509Err)
	}
}

func TestPlainEngineHTTPProxyMITMToUnknownAuthority(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine(WithMITM(localhost), proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, proxyServer)

	// WHEN
	_, err := tools.HTTPEcho(server.URL, "http proxy mitm to unknown authority")

	// THEN
	assert.Error(t, err, "Bad Request")
}

func TestPlainEngineHTTPProxyMITMToHTTP(t *testing.T) {
	// unusual, experimental combination, MITM is ON for HTTP server, so the client-proxy connection is upgraded to TLS
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	ca := newCA(t)
	serverCert := ca.SignLocalhost()
	tlsServerConfig := &tls.Config{
		Certificates: []tls.Certificate{*serverCert},
	}
	proxyServer := httptest.NewServer(
		proxy.NewEngine(WithMITM(localhost),
			proxy.WithServerConfig(func(string) (*tls.Config, error) {
				return tlsServerConfig, nil
			})))
	defer proxyServer.Close()
	tools := newHttpToolsWithRootCAs(t, proxyServer.URL, ca.RootCAs())

	// WHEN
	resp, err := tools.HTTPEcho(server.URL, "http proxy mitm to http")

	// THEN
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	}
}

func TestPlainEngineHTTPProxyMITMWithHandshakeTimeout(t *testing.T) {
	// GIVEN
	ca := newCA(t)
	serverCert := ca.SignLocalhost()
	server := httptest.NewUnstartedServer(newEchoServer(t))
	server.TLS = &tls.Config{
		GetConfigForClient: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
			// simulate TLS handshake timeout
			time.Sleep(200 * time.Millisecond)
			return nil, nil
		},
		Certificates: []tls.Certificate{*serverCert}, // pass some certificate, otherwise StartTLS will overwrite it
	}
	server.StartTLS()
	defer server.Close()

	proxyServer := httptest.NewServer(proxy.NewEngine(
		WithMITM(localhost),
		proxy.WithServerConfig(func(host string) (*tls.Config, error) {
			return &tls.Config{
				RootCAs: ca.RootCAs(),
			}, nil
		}),
		proxy.WithDialer(&net.Dialer{
			Timeout: 50 * time.Millisecond,
		})))
	defer proxyServer.Close()
	tools := newHttpTools(t, server.URL, proxyServer, server)

	// WHEN
	resp, respErr := tools.HTTPEcho(server.URL, "http proxy mitm to https with slow TLS handshake")

	// THEN
	assert.Error(t, respErr, "Gateway Timeout")
	assert.Nil(t, resp)
}

func TestPlainEngineHTTPProxyTunnelWithNetworkTimeout(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	slowDialer := &net.Dialer{
		ControlContext: func(ctx context.Context, network, address string, c syscall.RawConn) error {
			return context.DeadlineExceeded
		},
	}
	proxyServer := httptest.NewServer(proxy.NewEngine(
		proxy.WithDialer(slowDialer),
		proxy.WithServerConfig(func(_ string) (*tls.Config, error) {
			return server.TLS, nil
		}),
		proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, server)

	// WHEN
	resp, respErr := tools.HTTPEcho(server.URL, "http proxy tunnel to https with network timeout")

	// THEN
	assert.Error(t, respErr, "Gateway Timeout")
	assert.Nil(t, resp)
}

func TestPlainEngineHTTPProxyToHTTPWithNetworkTimeout(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	slowDialer := &net.Dialer{
		ControlContext: func(ctx context.Context, network, address string, c syscall.RawConn) error {
			return context.DeadlineExceeded
		},
	}
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithDialer(slowDialer), proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, proxyServer)

	// WHEN
	resp, respErr := tools.HTTPEcho(server.URL, "http proxy to http with network timeout")

	// THEN
	assert.NoError(t, respErr)
	if assert.NotNil(t, resp) {
		assert.Equal(t, resp.StatusCode, http.StatusGatewayTimeout)
	}
}

func TestPlainEngineHTTPProxyTunnelToHTTPSWithNetworkTimeout(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	slowDialer := &net.Dialer{
		ControlContext: func(ctx context.Context, network, address string, c syscall.RawConn) error {
			return context.DeadlineExceeded
		},
	}
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithDialer(slowDialer), proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, proxyServer)

	// WHEN
	resp, respErr := tools.HTTPEcho(server.URL, "http proxy to http with network timeout")

	// THEN
	assert.Error(t, respErr, "Gateway Timeout")
	assert.Nil(t, resp)
}

func TestPlainEngineHTTPProxyMITMToHTTPSWithKeepalive(t *testing.T) {
	// execute a sequence of HTTP requests via mitm proxy over the same TLS connection

	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine(WithTestServer(server, true)))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, proxyServer, server)

	// open TCP connection
	tcpConn, dialErr := net.Dial("tcp", proxyServer.Listener.Addr().String())
	require.NoError(t, dialErr)

	// request tunnel to the server
	_, writeConnectErr := tcpConn.Write([]byte(http.MethodConnect + " " + server.Listener.Addr().String() + " " + proxy.HTTP11 + "\r\n\r\n"))
	require.NoError(t, writeConnectErr)
	tcpBuff := bufio.NewReader(tcpConn)
	resp, readErr := http.ReadResponse(tcpBuff, nil)
	require.NoError(t, readErr)
	require.Equal(t, resp.StatusCode, http.StatusOK)

	// trigger TLS handshake from the client site
	tlsClientConfig := &tls.Config{RootCAs: tools.rootCAs, ServerName: localhost}
	tlsConn := tls.Client(tcpConn, tlsClientConfig)
	defer tools.Close(tlsConn)
	handshakeErr := tlsConn.Handshake()
	require.NoError(t, handshakeErr)

	// first request
	reqUrl, reqUrlErr := url.JoinPath(server.URL, "echo")
	require.NoError(t, reqUrlErr)

	createTestRequest := func(close bool) *http.Request {
		req := httptest.NewRequest(http.MethodGet, reqUrl, strings.NewReader("http proxy mitm to https with keepalive"))
		req.Close = close
		return req
	}

	firstReq := createTestRequest(false)
	firstWriteErr := firstReq.Write(tlsConn)
	require.NoError(t, firstWriteErr)

	tlsBuff := bufio.NewReader(tlsConn)
	firstResp, firstReadErr := http.ReadResponse(tlsBuff, nil)
	require.NoError(t, firstReadErr)
	require.Equal(t, firstResp.StatusCode, http.StatusOK)
	if assert.NotNil(t, firstResp.Body) {
		tools.Close(firstResp.Body)
	}

	// second request
	secondReq := createTestRequest(false)
	secondWriteErr := secondReq.Write(tlsConn)
	require.NoError(t, secondWriteErr)

	secondResp, secondReadErr := http.ReadResponse(tlsBuff, nil)
	require.NoError(t, secondReadErr)
	require.Equal(t, secondResp.StatusCode, http.StatusOK)
	if assert.NotNil(t, secondResp.Body) {
		tools.Close(secondResp.Body)
	}

	// third request
	thirdReq := createTestRequest(true)
	thirdWriteErr := thirdReq.Write(tlsConn)
	require.NoError(t, thirdWriteErr)

	thirdResp, thirdReadErr := http.ReadResponse(tlsBuff, nil)
	require.NoError(t, thirdReadErr)
	require.Equal(t, thirdResp.StatusCode, http.StatusOK)
	_, _ = io.Copy(io.Discard, tlsBuff) // discard all bytes

	// attempt to read from connection to confirm it is closed
	buffer := make([]byte, 16)
	readDeadlineErr := tlsConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	require.NoError(t, readDeadlineErr)
	nBytes, finalReadErr := tlsConn.Read(buffer)
	assert.Equal(t, nBytes, 0)
	assert.ErrorIs(t, finalReadErr, io.EOF)
}

func TestHTTPProxyToHTTPUsedMiddleware(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	mid := new(mockMiddleware)
	mid.On("Run", mock.Anything)
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithRule(&proxy.Rule{
		Action:   proxy.MITMAction,
		Handlers: []proxy.Handler{mid.Run},
	}, localhost)))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN-THEN
	tools.AssertHTTPEcho(server.URL, "http proxy to http")
	mid.AssertNumberOfCalls(t, "Run", 1)
}

func TestHTTPProxyMITMToHTTPSUsedMiddleware(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	mid := new(mockMiddleware)
	mid.On("Run", mock.Anything)
	proxyServer := httptest.NewServer(proxy.NewEngine(WithTestServer(server, true),
		proxy.WithRule(&proxy.Rule{
			Action:   proxy.MITMAction,
			Handlers: []proxy.Handler{mid.Run},
		}, localhost)))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, server)

	// WHEN-THEN
	tools.AssertHTTPEcho(server.URL, "http proxy mitm to http")
	mid.AssertNumberOfCalls(t, "Run", 2)
}

func TestHTTPProxyMITMToHTTPSWithPanicMiddleware(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	mid := func(c *proxy.Context) {
		panic("middleware malfunction")
	}
	proxyServer := httptest.NewServer(proxy.NewEngine(WithTestServer(server, true),
		proxy.WithRule(&proxy.Rule{
			Action:   proxy.MITMAction,
			Handlers: []proxy.Handler{mid},
		}, localhost), proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, server)

	// WHEN
	resp, respErr := tools.HTTPEcho(server.URL, "http proxy mitm to http")

	// THEN
	assert.Error(t, respErr, "Internal Server Error")
	assert.Nil(t, resp)
}

func TestHTTPProxyMITMToHTTPSWithConvertsNoResponseToInternalServerError(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	mid := func(c *proxy.Context) {
		c.Next()
		c.Response = nil
	}
	proxyServer := httptest.NewServer(proxy.NewEngine(WithTestServer(server, true),
		proxy.WithRule(&proxy.Rule{
			Action:   proxy.MITMAction,
			Handlers: []proxy.Handler{mid},
		}, localhost), proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, server)

	// WHEN
	resp, respErr := tools.HTTPEcho(server.URL, "http proxy mitm to http")

	// THEN
	assert.Error(t, respErr, "Internal Server Error")
	assert.Nil(t, resp)
}

func TestHTTPProxyToHTTPWithPanicMiddleware(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	handler := func(c *proxy.Context) {
		panic("middleware malfunction")
	}
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithRule(&proxy.Rule{
		Action:   proxy.MITMAction,
		Handlers: []proxy.Handler{handler},
	}, localhost), proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN
	resp, respErr := tools.HTTPEcho(server.URL, "http proxy mitm to http")

	// THEN
	assert.NoError(t, respErr)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestRuntimeWithoutHijackingConnections(t *testing.T) {
	// GIVEN
	engine := proxy.NewEngine(proxy.WithLogger(zerolog.Nop()))
	r := &http.Request{}
	w := new(mockResponseWriter)
	w.On("WriteHeader", http.StatusInternalServerError)

	// WHEN
	engine.ServeHTTP(w, r)

	// THEN
	w.AssertExpectations(t)
}

func TestHandlesMalformedHost(t *testing.T) {
	// GIVEN
	tools := newPipeTools()
	defer tools.Close()
	w := new(hijackingMockWriter)
	w.On("Hijack").Return(tools.serverConn, newEmptyReaderWriter(), nil)
	r := &http.Request{Host: "localhost"}
	engine := newEngineRunner(proxy.WithLogger(zerolog.Nop()))
	defer engine.Close()

	// WHEN
	engine.ServeHTTP(w, r)
	resp, readErr := tools.ReadResponse()

	// THEN
	assert.NoError(t, readErr)
	if assert.NotNil(t, resp) {
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
	w.AssertExpectations(t)
}

func TestHandlesClosedConnections(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	tools := newPipeTools()
	defer tools.Close()
	w := new(hijackingMockWriter)
	w.On("Hijack").Return(tools.serverConn, newEmptyReaderWriter(), nil)
	reqUrl, reqUrlErr := url.JoinPath(server.URL, "echo")
	require.NoError(t, reqUrlErr)
	r := httptest.NewRequest(http.MethodGet, reqUrl, nil)
	engine := newEngineRunner()

	// WHEN
	engine.ServeHTTP(w, r)
	resp, readErr := tools.ReadResponse()

	// THEN
	assert.NoError(t, readErr)
	if assert.NotNil(t, resp) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
	// trigger code path that detects closed client connections
	_ = tools.clientConn.Close()
	engine.Close()
	w.AssertExpectations(t)
}

func TestHandlesMalformedHTTPRequest(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	tools := newPipeTools()
	defer tools.Close()
	w := new(hijackingMockWriter)
	w.On("Hijack").Return(tools.serverConn, newEmptyReaderWriter(), nil)
	reqUrl, reqUrlErr := url.JoinPath(server.URL, "echo")
	require.NoError(t, reqUrlErr)
	r := httptest.NewRequest(http.MethodGet, reqUrl, nil)
	engine := newEngineRunner(proxy.WithLogger(zerolog.Nop()))
	defer engine.Close()

	// WHEN
	engine.ServeHTTP(w, r)
	resp, readErr := tools.ReadResponse()

	// THEN
	assert.NoError(t, readErr)
	if assert.NotNil(t, resp) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
	w.AssertExpectations(t)

	// WHEN
	_, writeErr := tools.clientConn.Write([]byte("malformed HTTP request\r\n\r\n"))

	// THEN
	require.NoError(t, writeErr)

	// WHEN
	resp, readErr = tools.ReadResponse()

	// THEN
	assert.NoError(t, readErr)
	if assert.NotNil(t, resp) {
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestPlainEngineHTTPProxyTunnelToHTTPSWithOtherError(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	customDialer := &net.Dialer{
		ControlContext: func(ctx context.Context, network, address string, c syscall.RawConn) error {
			// trigger handling of unexpected error
			return errors.New("other error")
		},
	}
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithDialer(customDialer),
		proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, server)

	// WHEN
	resp, respErr := tools.HTTPEcho(server.URL, "http proxy to http with network timeout")

	// THEN
	assert.Error(t, respErr, "Bad Gateway")
	assert.Nil(t, resp)
}

func TestPlainEngineHTTPProxyMITMToHTTPSWithOtherError(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	customDialer := &net.Dialer{
		ControlContext: func(ctx context.Context, network, address string, c syscall.RawConn) error {
			// trigger handling of unexpected error
			return errors.New("other error")
		},
	}
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithDialer(customDialer), WithTestServer(server, true),
		proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, server)

	// WHEN
	resp, respErr := tools.HTTPEcho(server.URL, "http proxy to http with network timeout")

	// THEN
	assert.Error(t, respErr, "Bad Gateway")
	assert.Nil(t, resp)
}

func TestPlainEngineHTTPProxyMITMToHTTPSWithDialTimeout(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	dialErrors := []error{nil, errors.New("custom error")}
	index := 0
	customDialer := &net.Dialer{
		ControlContext: func(ctx context.Context, network, address string, c syscall.RawConn) error {
			var err error
			if index < len(dialErrors) {
				err = dialErrors[index]
				index++
			}
			return err
		},
	}
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithDialer(customDialer),
		WithTestServer(server, true),
		proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN
	resp, respErr := tools.HTTPEcho("https://"+server.Listener.Addr().String(), "http proxy to http with network timeout")

	// THEN
	assert.EqualError(t, respErr, "Bad Gateway")
	assert.Nil(t, resp)
}

func TestPlainEngineHTTPProxyMITMToHTTPSWithCustomError(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	dialErrors := []error{nil, context.DeadlineExceeded}
	index := 0
	customDialer := &net.Dialer{
		ControlContext: func(ctx context.Context, network, address string, c syscall.RawConn) error {
			var err error
			if index < len(dialErrors) {
				err = dialErrors[index]
				index++
			}
			return err
		},
	}
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithDialer(customDialer),
		WithTestServer(server, true),
		proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN
	resp, respErr := tools.HTTPEcho("https://"+server.Listener.Addr().String(), "http proxy to http with network timeout")

	// THEN
	assert.EqualError(t, respErr, "Gateway Timeout")
	assert.Nil(t, resp)
}

func TestPlainEngineHTTPProxyMITMToHTTPSWithTLSTimeout(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	customDialer := &net.Dialer{
		ControlContext: func(ctx context.Context, network, address string, c syscall.RawConn) error {
			return context.DeadlineExceeded
		},
	}
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithDialer(customDialer), WithTestServer(server, true), proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN
	resp, respErr := tools.HTTPEcho("https://"+server.Listener.Addr().String(), "http proxy mitm to http")

	// THEN
	assert.EqualError(t, respErr, "Gateway Timeout")
	assert.Nil(t, resp)
}

func TestPlainEngineHTTPProxyMITMToHTTPWithTLSConfigError(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine(
		proxy.WithServerConfig(func(_ string) (*tls.Config, error) {
			return nil, errors.New("custom error")
		}),
		WithTestServer(server, true),
		proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN
	resp, respErr := tools.HTTPEcho("https://"+server.Listener.Addr().String(), "http proxy mitm to http")

	// THEN
	assert.EqualError(t, respErr, "Internal Server Error")
	assert.Nil(t, resp)
}

func TestPlainEngineHTTPProxyMITMToHTTPSWithTLSConfigError(t *testing.T) {
	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine(
		WithTestServer(server, true),
		proxy.WithServerConfig(func(_ string) (*tls.Config, error) {
			return nil, errors.New("custom error")
		}),
		proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, server)

	// WHEN
	resp, respErr := tools.HTTPEcho(server.URL, "http proxy mitm to http")

	// THEN
	assert.EqualError(t, respErr, "Internal Server Error")
	assert.Nil(t, resp)
}

func TestPlainEngineHTTPProxyTCPServer(t *testing.T) {
	// GIVEN
	server := newTCPServer(t)
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine(proxy.WithLogger(zerolog.Nop())))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL)

	// WHEN
	resp, respErr := tools.HTTPEcho(server.URL, "http proxy to tcp")

	// THEN
	assert.NoError(t, respErr)
	if assert.NotNil(t, resp) {
		assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
	}
}

func TestPlainEngineHTTPProxyMITMToWSSPlain(t *testing.T) {
	// execute a sequence of HTTP requests via mitm proxy over the same TLS connection

	// GIVEN
	server := httptest.NewTLSServer(newEchoServer(t))
	defer server.Close()
	proxyServer := httptest.NewServer(proxy.NewEngine(WithTestServer(server, true)))
	defer proxyServer.Close()
	tools := newHttpTools(t, proxyServer.URL, proxyServer, server)

	// open TCP connection
	tcpConn, dialErr := net.Dial("tcp", proxyServer.Listener.Addr().String())
	require.NoError(t, dialErr)

	// request tunnel to the server
	_, writeConnectErr := tcpConn.Write([]byte(http.MethodConnect + " " + server.Listener.Addr().String() + " " + proxy.HTTP11 + "\r\n\r\n"))
	require.NoError(t, writeConnectErr)
	tcpBuff := bufio.NewReader(tcpConn)
	resp, readErr := http.ReadResponse(tcpBuff, nil)
	require.NoError(t, readErr)
	require.Equal(t, resp.StatusCode, http.StatusOK)

	// trigger TLS handshake from the client site
	tlsClientConfig := &tls.Config{RootCAs: tools.rootCAs, ServerName: localhost}
	tlsConn := tls.Client(tcpConn, tlsClientConfig)
	defer tools.Close(tlsConn)
	handshakeErr := tlsConn.Handshake()
	require.NoError(t, handshakeErr)

	reqUrl, reqUrlErr := url.JoinPath(server.URL, "echo")
	require.NoError(t, reqUrlErr)
	req := httptest.NewRequest(http.MethodGet, reqUrl, nil)
	req.Header.Add("Connection", "upgrade")
	req.Header.Add("Upgrade", "websocket")
	// Sec-Websocket-Version must be set to 13
	req.Header.Add("Sec-Websocket-Version", "13")
	// Sec-Websocket-Key must be a base64 encoded string of length 16
	req.Header.Add("Sec-Websocket-Key", base64.StdEncoding.EncodeToString([]byte("1234567890ABCDEF")))

	writeErr := req.Write(tlsConn)
	require.NoError(t, writeErr)

	tlsBuff := bufio.NewReader(tlsConn)
	firstResp, firstReadErr := http.ReadResponse(tlsBuff, nil)
	require.NoError(t, firstReadErr)
	require.Equal(t, http.StatusSwitchingProtocols, firstResp.StatusCode)
}

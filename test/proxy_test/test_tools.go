/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy_test

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const localhost = "127.0.0.1"

func newEmptyReaderWriter() *bufio.ReadWriter {
	return bufio.NewReadWriter(
		bufio.NewReader(bytes.NewReader([]byte{})),
		bufio.NewWriter(io.Discard))
}

type testTools struct {
	*testing.T
	proxy     *url.URL
	rootCAs   *x509.CertPool
	transport *http.Transport
}

func newHttpTools(t *testing.T, proxyUrl string, servers ...*httptest.Server) *testTools {
	rootCAs := x509.NewCertPool()
	for _, server := range servers {
		cert := server.Certificate()
		if cert != nil {
			rootCAs.AddCert(cert)
		}
	}

	return newHttpToolsWithRootCAs(t, proxyUrl, rootCAs)
}

func newHttpToolsWithRootCAs(t *testing.T, proxyUrl string, rootCAs *x509.CertPool) *testTools {
	parsedProxyUrl, parseErr := url.Parse(proxyUrl)
	if parseErr != nil {
		t.Fatalf("test tools: failed to parse url %s: %v", proxyUrl, parseErr)
	}

	transport := &http.Transport{
		Proxy: func(request *http.Request) (*url.URL, error) {
			return parsedProxyUrl, nil
		},
		TLSClientConfig: &tls.Config{RootCAs: rootCAs},
	}

	return &testTools{
		T:         t,
		proxy:     parsedProxyUrl,
		rootCAs:   rootCAs,
		transport: transport,
	}
}

func (t *testTools) Proxy(r *http.Request) (*url.URL, error) {
	return t.proxy, nil
}

func (t *testTools) HTTPEcho(serverUrl, message string) (*http.Response, error) {
	parsedServerUrl, parseErr := url.Parse(serverUrl)
	if parseErr != nil {
		t.Fatalf("test tools: failed to parse url %s: %v", serverUrl, parseErr)
	}

	reqUrl := parsedServerUrl.JoinPath("/echo").String()
	req, reqErr := http.NewRequest(http.MethodGet, reqUrl, strings.NewReader(message))
	if reqErr != nil {
		t.Fatalf("test tools: failed to create GET %s request: %v", reqUrl, reqErr)
	}

	return t.transport.RoundTrip(req)
}

func (t *testTools) Close(closer io.Closer) {
	closeErr := closer.Close()
	if closeErr != nil {
		t.Logf("test tools: failed to close: %v", closeErr)
	}
}

func (t *testTools) AssertHTTPEcho(serverUrl, message string) {
	resp, err := t.HTTPEcho(serverUrl, message)

	assert.NoErrorf(t, err, "test tools: failed to echo %s: %v", serverUrl, err)
	if !assert.NotNil(t, resp) {
		return
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	if !assert.NotNil(t, resp.Body) {
		return
	}

	defer t.Close(resp.Body)
	content, readErr := io.ReadAll(resp.Body)
	if assert.NoError(t, readErr) {
		assert.Equal(t, message, string(content))
	}
}

func (t *testTools) DialWebsocket(serverUrl string) (*websocket.Conn, *http.Response, error) {
	parsedServerUrl, parseErr := url.Parse(serverUrl)
	if parseErr != nil {
		t.Fatalf("test tools: failed to parse url %s: %v", serverUrl, parseErr)
	}

	reqUrl := parsedServerUrl.JoinPath("echo")
	if parsedServerUrl.Scheme == "https" {
		reqUrl.Scheme = "wss"
	} else {
		reqUrl.Scheme = "ws"
	}

	dialer := &websocket.Dialer{
		Proxy: t.Proxy,
		TLSClientConfig: &tls.Config{
			RootCAs: t.rootCAs,
		},
	}

	return dialer.Dial(reqUrl.String(), nil)
}

func (t *testTools) AssertWebsocketEcho(serverUrl, message string) {
	// WHEN
	wsConn, resp, dialErr := t.DialWebsocket(serverUrl)

	// THEN
	if !assert.NoError(t, dialErr) {
		return
	}
	if !assert.NotNil(t, wsConn) {
		return
	}
	defer t.Close(wsConn)
	if !assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode) {
		return
	}

	// WHEN
	writeErr := wsConn.WriteMessage(websocket.TextMessage, []byte(message))

	// THEN
	if !assert.NoError(t, writeErr) {
		return
	}

	// WHEN
	msgType, rawMsg, readErr := wsConn.ReadMessage()

	// THEN
	if !assert.NoError(t, readErr) {
		return
	}
	assert.Equal(t, websocket.TextMessage, msgType)
	assert.Equal(t, message, string(rawMsg))
}

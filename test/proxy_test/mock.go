/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy_test

import (
	"bufio"
	"github.com/pmateusz/glove/pkg/proxy"
	"github.com/stretchr/testify/mock"
	"net"
	"net/http"
)

type mockMiddleware struct {
	mock.Mock
}

func (m *mockMiddleware) Run(ctx *proxy.Context) {
	m.Called(ctx)

	ctx.Next()
}

type mockResponseWriter struct {
	mock.Mock
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *mockResponseWriter) Header() http.Header {
	args := m.Called()
	return args.Get(0).(http.Header)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.Called(statusCode)
}

type hijackingMockWriter struct {
	mockResponseWriter
}

func (m *hijackingMockWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	args := m.Called()
	return args.Get(0).(net.Conn), args.Get(1).(*bufio.ReadWriter), args.Error(2)
}

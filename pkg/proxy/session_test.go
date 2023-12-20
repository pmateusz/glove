/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy

import (
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"
)

type mockConn struct {
	mock.Mock
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *mockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockConn) LocalAddr() net.Addr {
	args := m.Called()
	addr := args.Get(0)
	if addr == nil {
		return nil
	}
	return addr.(net.Addr)
}

func (m *mockConn) RemoteAddr() net.Addr {
	args := m.Called()
	addr := args.Get(0)
	if addr == nil {
		return nil
	}
	return addr.(net.Addr)
}

func (m *mockConn) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

type addressDouble struct {
	net  string
	addr string
}

func newAddress(net, addr string) addressDouble {
	return addressDouble{net, addr}
}

func (a addressDouble) Network() string {
	return a.net
}

func (a addressDouble) String() string {
	return a.addr
}

func okMiddleware(c *Context) {
	c.Response = newHTTP11Response(http.StatusOK, nil)
}

func newMockConnWithWriteError(err error) *mockConn {
	conn := new(mockConn)
	conn.On("Write", mock.Anything).Return(0, err)
	conn.On("RemoteAddr").Return(newAddress("tcp", "127.0.0.1:1"))
	conn.On("LocalAddr").Return(newAddress("tcp", "127.0.0.1:2"))
	return conn
}

func TestHandleEOFWhileWritingResponse(t *testing.T) {
	// GIVEN
	conn := newMockConnWithWriteError(io.EOF)
	s := &session{
		rule:       &Rule{Handlers: []Handler{okMiddleware}},
		clientConn: conn,
		tools:      newNetTools(zerolog.New(zerolog.NewTestWriter(t))),
	}
	r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)

	// WHEN
	s.handle(r)

	// THEN
	assert.True(t, s.close)
	conn.AssertExpectations(t)
}

func TestHandleSyscallErrorWhileWritingResponse(t *testing.T) {
	// GIVEN
	conn := new(mockConn)
	conn.On("Write", mock.Anything).Return(0, syscall.EPIPE)
	s := &session{
		rule:       &Rule{Handlers: []Handler{okMiddleware}},
		clientConn: conn,
		tools:      newNetTools(zerolog.New(zerolog.NewTestWriter(t))),
	}
	r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)

	// WHEN
	s.handle(r)

	// THEN
	assert.True(t, s.close)
	conn.AssertExpectations(t)
}

func TestDetectsClosedServerConnection(t *testing.T) {
	// GIVEN
	conn := newMockConnWithWriteError(syscall.EPIPE)
	s := &session{
		rule:       &Rule{Action: TunnelAction, Handlers: []Handler{okMiddleware}},
		serverConn: conn,
		tools:      newNetTools(zerolog.New(zerolog.NewTestWriter(t))),
	}
	r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	c := &Context{
		Request: r,
		s:       s,
	}

	// WHEN
	s.execute(c)

	// THEN
	assert.True(t, s.close)
	conn.AssertExpectations(t)
}

/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package acl

import (
	"github.com/stretchr/testify/mock"
	"net"
	"time"
)

type mockListener struct {
	mock.Mock
}

func (m *mockListener) Accept() (net.Conn, error) {
	args := m.Called()
	conn, _ := args.Get(0).(net.Conn)
	return conn, args.Error(1)
}

func (m *mockListener) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockListener) Addr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

type mockACL struct {
	mock.Mock
}

func (m *mockACL) Allowed(ip net.IP) bool {
	args := m.Called(ip)
	return args.Bool(0)
}

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
	return args.Get(0).(net.Addr)
}

func (m *mockConn) RemoteAddr() net.Addr {
	args := m.Called()
	conn, _ := args.Get(0).(net.Addr)
	return conn
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

type mockAddr struct {
	mock.Mock
}

func (m *mockAddr) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockAddr) Network() string {
	args := m.Called()
	return args.String(0)
}

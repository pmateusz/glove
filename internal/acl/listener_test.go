/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package acl

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net"
	"testing"
)

func TestAllowsConnection(t *testing.T) {
	// GIVEN
	addr := new(mockAddr)
	addr.On("String").Return("127.0.0.1:1234")
	tcpConn := new(mockConn)
	tcpConn.On("RemoteAddr").Return(addr)
	tcpListener := new(mockListener)
	tcpListener.On("Accept").Return(tcpConn, nil).Once()
	acl := new(mockACL)
	acl.On("Allowed", mock.Anything).Return(true)
	listener := WrapListener(zerolog.Nop(), acl, tcpListener)

	// WHEN
	conn, acceptErr := listener.Accept()

	// THEN
	assert.NoError(t, acceptErr)
	assert.Equal(t, tcpConn, conn)
}

func TestClosesConnectionForForbiddenIp(t *testing.T) {
	// GIVEN
	blockedIp := net.ParseIP("123.123.123.123")
	addr := new(mockAddr)
	addr.On("String").Return(blockedIp.String() + ":1234")
	tcpConn := new(mockConn)
	tcpConn.On("RemoteAddr").Return(addr)
	tcpConn.On("Close").Return(nil).Once()
	tcpListener := new(mockListener)
	tcpListener.On("Accept").Return(tcpConn, nil).Once()
	tcpListener.On("Accept").Return(nil, errors.New("listener: stop")).Once()
	acl := new(mockACL)
	acl.On("Allowed",
		mock.MatchedBy(func(ip net.IP) bool {
			return ip.Equal(blockedIp)
		})).Return(false).Once()
	listener := WrapListener(zerolog.Nop(), acl, tcpListener)

	// WHEN
	conn, err := listener.Accept()

	// THEN
	assert.Error(t, err, "listener: stop")
	assert.Nil(t, conn)
	tcpConn.AssertExpectations(t)
}

func TestHandlesCloseErrorForForbiddenIp(t *testing.T) {
	// GIVEN
	addr := new(mockAddr)
	addr.On("String").Return("127.0.0.1:1234")
	tcpConn := new(mockConn)
	tcpConn.On("RemoteAddr").Return(addr)
	tcpConn.On("Close").Return(errors.New("mask: close error")).Once()
	tcpListener := new(mockListener)
	tcpListener.On("Accept").Return(tcpConn, nil).Once()
	tcpListener.On("Accept").Return(nil, errors.New("listener: stop")).Once()
	acl := new(mockACL)
	acl.On("Allowed", mock.Anything).Return(false).Once()
	listener := WrapListener(zerolog.Nop(), acl, tcpListener)

	// WHEN
	conn, acceptErr := listener.Accept()

	// THEN
	assert.Error(t, acceptErr, "listener: stop")
	assert.Nil(t, conn)
	acl.AssertExpectations(t)
	tcpConn.AssertExpectations(t)
}

func TestHandlesIPLookupError(t *testing.T) {
	// GIVEN
	addr := new(mockAddr)
	addr.On("String").Return("127.0.0.1")
	tcpConn := new(mockConn)
	tcpConn.On("RemoteAddr").Return(addr)
	tcpConn.On("Close").Return(nil).Once()
	tcpListener := new(mockListener)
	tcpListener.On("Accept").Return(tcpConn, nil).Once()
	listener := WrapListener(zerolog.Nop(), nil, tcpListener)

	// WHEN
	conn, acceptErr := listener.Accept()

	// THEN
	assert.EqualError(t, acceptErr, "address 127.0.0.1: missing port in address")
	assert.Nil(t, conn)
	tcpConn.AssertExpectations(t)
}

func TestHandlesNoRemoteAddressError(t *testing.T) {
	// GIVEN
	tcpConn := new(mockConn)
	tcpConn.On("RemoteAddr").Return(nil)
	tcpConn.On("Close").Return(nil).Once()
	tcpListener := new(mockListener)
	tcpListener.On("Accept").Return(tcpConn, nil).Once()
	listener := WrapListener(zerolog.Nop(), nil, tcpListener)

	// WHEN
	conn, acceptErr := listener.Accept()

	// THEN
	assert.Equal(t, acceptErr, ErrNoRemoteAddress)
	assert.Nil(t, conn)
	tcpConn.AssertExpectations(t)
}

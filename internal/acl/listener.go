/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package acl

import (
	"errors"
	"github.com/rs/zerolog"
	"net"
)

var ErrNoRemoteAddress = errors.New("acl: no remote address")

type Listener struct {
	log      zerolog.Logger
	listener net.Listener
	acl      ACL
}

func WrapListener(log zerolog.Logger, acl ACL, listener net.Listener) *Listener {
	return &Listener{
		log:      log,
		acl:      acl,
		listener: listener,
	}
}

func (l *Listener) Accept() (net.Conn, error) {
	for {
		conn, acceptErr := l.listener.Accept()
		if acceptErr != nil {
			return nil, acceptErr
		}

		ip, lookupErr := netConnLookup(conn)
		if lookupErr != nil {
			l.closeConn(conn, ip)
			return nil, lookupErr
		}

		if l.acl.Allowed(ip) {
			return conn, nil
		}

		// IP address is not allowed, close the connection and resume listening.
		// Error handling has to be completed here. Returning an error to the callee
		// will shut down the http.Server.
		l.log.Info().IPAddr("remoteAddr", ip).Msg("block-ip")
		l.closeConn(conn, ip)
	}
}

func (l *Listener) closeConn(conn net.Conn, ip net.IP) {
	if closeErr := conn.Close(); closeErr != nil {
		event := l.log.Err(closeErr)
		if !ip.IsUnspecified() {
			event.IPAddr("remoteAddr", ip)
		}
		event.Msg("close-network-connection")
	}
}

func (l *Listener) Close() error {
	return l.listener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

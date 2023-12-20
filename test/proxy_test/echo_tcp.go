/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy_test

import (
	"fmt"
	"io"
	"net"
	"testing"
)

type echoTCPServer struct {
	t *testing.T
	l net.Listener

	URL string
}

func newTCPServer(t *testing.T) *echoTCPServer {
	server := &echoTCPServer{
		t: t,
	}
	server.Start()
	return server
}

func (s *echoTCPServer) Start() {
	var listenErr error
	s.l, listenErr = net.Listen("tcp4", "127.0.0.1:0")
	if listenErr != nil {
		panic(fmt.Sprintf("tcp server: failed to start: %v", listenErr))
	}

	s.URL = "http://" + s.l.Addr().String()

	go func() {
		for {
			conn, acceptErr := s.l.Accept()
			if acceptErr != nil {
				s.t.Logf("tcp server: failed to accept the connection: %v", acceptErr)
				return
			}

			go s.handle(conn)
		}
	}()
}

func (s *echoTCPServer) handle(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	_, copyErr := io.Copy(conn, conn)
	if copyErr != nil {
		s.t.Logf("tcp server: failed to copy bytes: %v", copyErr)
	}
}

func (s *echoTCPServer) Close() {
	_ = s.l.Close()
}

/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy_test

import (
	"bufio"
	"net"
	"net/http"
)

type pipeTools struct {
	clientConn   net.Conn
	clientReader *bufio.Reader
	serverConn   net.Conn
}

func newPipeTools() *pipeTools {
	clientConn, serverConn := net.Pipe()
	return &pipeTools{
		clientConn:   clientConn,
		clientReader: bufio.NewReader(clientConn),
		serverConn:   serverConn,
	}
}

func (t *pipeTools) ReadResponse() (*http.Response, error) {
	return http.ReadResponse(t.clientReader, nil)
}

func (t *pipeTools) Close() {
	_ = t.clientConn.Close()
	_ = t.serverConn.Close()
}

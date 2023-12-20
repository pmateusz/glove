/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy_test

import (
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"testing"
)

type echoServer struct {
	t *testing.T
	u websocket.Upgrader
}

func newEchoServer(t *testing.T) *echoServer {
	return &echoServer{t, websocket.Upgrader{}}
}

func (s *echoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/echo" {
		http.NotFound(w, r)
		return
	}

	if websocket.IsWebSocketUpgrade(r) {
		conn, upgradeErr := s.u.Upgrade(w, r, nil)
		if upgradeErr != nil {
			s.t.Logf("server: upgrade to web socket protocol failed: %v", upgradeErr)
			return
		}

		defer func() {
			closeErr := conn.Close()
			if closeErr != nil {
				s.t.Logf("server: failed to close connection: %v", closeErr)
			}
		}()

		for {
			msgType, msg, readErr := conn.ReadMessage()
			if readErr != nil {
				s.t.Logf("server: failed to read web socket message: %v", readErr)
				break
			}
			s.t.Logf("server: received web socker message: %s", msg)

			if writeErr := conn.WriteMessage(msgType, msg); writeErr != nil {
				s.t.Logf("server: failed to write message: %v", writeErr)
				break
			}
		}
	} else {
		w.WriteHeader(http.StatusOK)
		_, writeBodyErr := io.Copy(w, r.Body)
		if writeBodyErr != nil {
			s.t.Logf("server: failed to write response: %v", writeBodyErr)
		}
	}
}

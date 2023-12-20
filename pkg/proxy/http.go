/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy

import (
	"golang.org/x/net/http/httpguts"
	"net/http"
)

func newHTTP10ConnectionEstablished(r *http.Request) *http.Response {
	return &http.Response{ProtoMajor: 1, ProtoMinor: 0, StatusCode: http.StatusOK, Status: "200 Connection established", Request: r}
}

func newHTTP11Response(statusCode int, r *http.Request) *http.Response {
	return &http.Response{ProtoMajor: 1, ProtoMinor: 1, StatusCode: statusCode, Request: r}
}

func isWebsocketUpgrade(r *http.Request) bool {
	return httpguts.HeaderValuesContainsToken(r.Header["Connection"], "upgrade") && httpguts.HeaderValuesContainsToken(r.Header["Upgrade"], "websocket")
}

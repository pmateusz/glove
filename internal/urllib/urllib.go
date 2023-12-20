/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package urllib

import (
	"net"
	"net/http"
	"strings"
)

func ClientIP(r *http.Request) net.IP {
	remoteIP := RemoteIp(r)
	if remoteIP == nil {
		return nil
	}

	for _, headerName := range []string{"X-Forwarded-For", "X-Real-IP"} {
		ip := parseIpHeader(r.Header.Get(headerName))
		if ip != nil {
			return ip
		}
	}

	return remoteIP
}

func RemoteIp(r *http.Request) net.IP {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return nil
	}
	return net.ParseIP(ip)
}

func parseIpHeader(header string) net.IP {
	if header == "" {
		return nil
	}

	items := strings.Split(header, ",")
	for i := len(items) - 1; i >= 0; i-- {
		ipStr := strings.TrimSpace(items[i])
		ip := net.ParseIP(ipStr)
		if ip != nil {
			return ip
		}
	}

	return nil
}

func RemoveProxyHeaders(r *http.Request) {
	r.RequestURI = "" // this must be reset when serving a request with the client

	// If no Accept-Encoding header exists, Transport will add the headers it can accept
	// and would wrap the response body with the relevant reader.
	r.Header.Del("Accept-Encoding")

	// curl can add that, see
	// https://jdebp.eu./FGA/web-proxy-connection-header.html
	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")

	// Connection, Authenticate and Authorization are single hop Header:
	// http://www.w3.org/Protocols/rfc2616/rfc2616.txt
	// 14.10 Connection
	//   The Connection general-header field allows the sender to specify
	//   options that are desired for that particular connection and MUST NOT
	//   be communicated by proxies over further connections.

	// When server reads http request it sets req.Close to true if
	// "Connection" header contains "close".
	// https://github.com/golang/go/blob/master/src/net/http/request.go#L1080
	// Later, transfer.go adds "Connection: close" back when req.Close is true
	// https://github.com/golang/go/blob/master/src/net/http/transfer.go#L275
	// That's why tests that checks "Connection: close" removal fail
	if r.Header.Get("Connection") == "close" {
		r.Close = false
	}
	r.Header.Del("Connection")
}

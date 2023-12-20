/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package acl

import "net"

func netConnLookup(conn net.Conn) (net.IP, error) {
	remoteAddr := conn.RemoteAddr()
	if remoteAddr == nil {
		return nil, ErrNoRemoteAddress
	}

	host, _, err := net.SplitHostPort(remoteAddr.String())
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(host)
	return ip, nil
}

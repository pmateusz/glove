/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy

import (
	"errors"
	"github.com/rs/zerolog"
	"net"
	"os"
)

func withOpError(event *zerolog.Event, networkErr *net.OpError) *zerolog.Event {
	event = event.Str("op", networkErr.Op).
		Str("net", networkErr.Net)

	if networkErr.Source != nil {
		event = event.Str("source", networkErr.Source.String())
	}

	if networkErr.Addr != nil {
		event = event.Str("addr", networkErr.Addr.String())
	}

	var syscallErr *os.SyscallError
	if errors.As(networkErr.Err, &syscallErr) {
		return event.Str("syscall", syscallErr.Syscall).Err(syscallErr.Err)
	}

	return event.Err(networkErr.Err)
}

func withConn(event *zerolog.Event, conn net.Conn) *zerolog.Event {
	return event.Str("remoteAddr", conn.RemoteAddr().String()).
		Str("localAddr", conn.LocalAddr().String())
}

func withRemoteAddr(event *zerolog.Event, remoteAddr string) *zerolog.Event {
	return event.Str("remoteAddr", remoteAddr)
}

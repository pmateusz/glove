/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy

import (
	"bytes"
	"errors"
	"github.com/rs/zerolog"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
)

const (
	HTTP11 = "HTTP/1.1"
	HTTP10 = "HTTP/1.0"
)

var (
	errNotSupported = errors.New("runtime: not supported")
)

type netTools struct {
	logger zerolog.Logger
}

func newNetTools(logger zerolog.Logger) *netTools {
	return &netTools{
		logger: logger,
	}
}

func (t *netTools) Copy(dest net.Conn, source net.Conn) {
	if nBytes, err := io.Copy(dest, source); err != nil {
		t.logger.Info().
			Err(err).
			Int64("bytesWritten", nBytes).
			Str("sourceAddr", source.RemoteAddr().String()).
			Str("destAddr", dest.RemoteAddr().String()).
			Msg("copy")
	}
}

func (t *netTools) Pipe(left, right net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		t.Copy(left, right)
	}()

	go func() {
		defer wg.Done()
		t.Copy(right, left)
	}()

	wg.Wait()
}

func (t *netTools) Hijack(w http.ResponseWriter) (net.Conn, error) {
	hij, ok := w.(http.Hijacker)
	if !ok {
		return nil, errNotSupported
	}

	conn, _, err := hij.Hijack()
	return conn, err
}

func (t *netTools) CloseConn(conn net.Conn) {
	closeErr := conn.Close()
	if closeErr != nil {
		withConn(t.logger.Info(), conn).Err(closeErr).Msg("close")
	}
}

func (t *netTools) CloseBody(resp *http.Response) {
	if resp.Body == nil {
		return
	}

	if err := resp.Body.Close(); err != nil {
		t.logger.Info().Err(err).Msg("close")
	}
}

func (t *netTools) WriteHTTP11Status(conn net.Conn, code int) {
	writeErr := t.WriteHTTPStatusE(conn, HTTP11, code, http.StatusText(code))
	if writeErr != nil {
		withConn(t.logger.Info(), conn).Err(writeErr).Msg("write")
	}
}

func (t *netTools) WriteHTTPStatusE(conn net.Conn, protocol string, code int, statusText string) error {
	b := bytes.NewBuffer(make([]byte, 0, 32))
	b.WriteString(protocol)
	b.WriteByte(' ')
	b.WriteString(strconv.Itoa(code))
	b.WriteByte(' ')
	b.WriteString(statusText)
	b.Write([]byte("\r\n\r\n"))

	_, err := b.WriteTo(conn)
	return err
}

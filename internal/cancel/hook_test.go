/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package cancel

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"
)

type mockCloser struct {
	mock.Mock
}

func (c *mockCloser) Close() error {
	args := c.Called()
	return args.Error(0)
}

func testClosesAfterSignal(t *testing.T, s syscall.Signal) {
	// GIVEN
	hook := NewHook(context.Background(), zerolog.Nop())
	var closer mockCloser
	closer.On("Close").Once().Return(nil)
	hook.Register("closer", &closer)

	// WHEN
	hook.Start()
	hook.signals <- s

	// THEN
	<-hook.Done()
	closer.AssertExpectations(t)
}

func TestClosesAfterSIGTERM(t *testing.T) {
	testClosesAfterSignal(t, syscall.SIGTERM)
}

func TestClosesAfterSIGINT(t *testing.T) {
	testClosesAfterSignal(t, syscall.SIGINT)
}

func TestClosesAfterContextCancel(t *testing.T) {
	// GIVEN
	var closer mockCloser
	closer.On("Close").Once().Return(nil)
	ctx, cancel := context.WithCancel(context.TODO())
	hook := NewHook(ctx, zerolog.Nop())
	hook.Register("closer", &closer)

	// WHEN
	hook.Start()
	cancel()

	// THEN
	<-hook.Done()
	closer.AssertExpectations(t)
}

type httpHandlerDouble struct{}

func (h *httpHandlerDouble) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestShutdownHttpServer(t *testing.T) {
	// GIVEN
	server := httptest.NewServer(&httpHandlerDouble{})
	defer server.Close()
	shutdown := make(chan struct{}, 1)
	server.Config.RegisterOnShutdown(func() {
		close(shutdown)
	})

	ctx, cancel := context.WithCancel(context.TODO())
	hook := NewHook(ctx, zerolog.Nop())
	hook.Register("server", WrapServer(server.Config, time.Second))

	// WHEN
	hook.Start()
	cancel()

	// THEN
	<-hook.Done()
	<-shutdown
}

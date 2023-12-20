/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package cancel

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Hook calls Close method on registered io.Closer instances in response to SIGINT, SIGTERM signals or whenever its
// Context expires or is cancelled
type Hook struct {
	log     zerolog.Logger
	mu      sync.Mutex
	closers map[string]io.Closer

	ctx     context.Context
	signals chan os.Signal
	done    chan struct{}
}

func (h *Hook) Start() {
	signal.Notify(h.signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-h.signals:
			h.log.Info().Str("name", sig.String()).Msg("signal")
		case <-h.ctx.Done():
			h.log.Info().Str("state", "expired-or-cancelled").Msg("context")
		}

		signal.Stop(h.signals)
		close(h.signals)

		h.cancel()
		close(h.done)
	}()
}

func (h *Hook) cancel() {
	h.log.Info().Str("state", "started").Msg("shutdown")

	var wg sync.WaitGroup
	h.mu.Lock()
	for resource, closer := range h.closers {
		wg.Add(1)
		go func(resource string, closer io.Closer) {
			defer wg.Done()

			err := closer.Close()
			if err != nil {
				h.log.Err(err).Str("resource", resource).Msg("close")
			}
		}(resource, closer)
	}
	h.mu.Unlock()
	wg.Wait()

	h.log.Info().Str("state", "completed").Msg("shutdown")
}

func (h *Hook) Register(resourceName string, closer io.Closer) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.closers[resourceName] = closer
}

func (h *Hook) Done() <-chan struct{} {
	return h.done
}

type serverCloser struct {
	server   *http.Server
	deadline time.Duration
}

func (c *serverCloser) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.deadline)
	defer cancel()

	err := c.server.Shutdown(ctx)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func NewHook(ctx context.Context, log zerolog.Logger) *Hook {
	return &Hook{
		log:     log,
		mu:      sync.Mutex{},
		closers: map[string]io.Closer{},
		ctx:     ctx,
		signals: make(chan os.Signal, 1),
		done:    make(chan struct{}, 1),
	}
}

func WrapServer(server *http.Server, deadline time.Duration) io.Closer {
	return &serverCloser{
		server:   server,
		deadline: deadline,
	}
}

/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package limiter

import (
	"github.com/pmateusz/glove/pkg/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/time/rate"
	"net/http"
	"testing"
	"time"
)

type mockHandler struct {
	mock.Mock
}

func (h *mockHandler) Handle(c *proxy.Context) {
	h.Called(c)
}

func TestPassRequestWithoutThrottling(t *testing.T) {
	// GIVEN
	handler := new(mockHandler)
	handler.On("Handle", mock.Anything).Once()
	ctx := proxy.NewTestOnlyContext(handler.Handle)
	limiter := &Handler{rate.NewLimiter(1, 1)}

	// WHEN
	limiter.Handle(ctx)

	// THEN
	handler.AssertExpectations(t)
}

func TestCanThrottleRequest(t *testing.T) {
	// GIVEN
	handler := new(mockHandler)
	handler.On("Handle", mock.Anything).Once()
	ctx := proxy.NewTestOnlyContext(handler.Handle)
	innerLimiter := rate.NewLimiter(rate.Every(time.Millisecond), 1)
	limiter := &Handler{innerLimiter}

	// WHEN
	start := time.Now()
	innerLimiter.Allow()
	limiter.Handle(ctx)
	end := time.Now()

	// THEN
	handler.AssertExpectations(t)
	assert.LessOrEqual(t, time.Millisecond, end.Sub(start))
}

func TestCanRejectInfeasibleRequest(t *testing.T) {
	// GIVEN
	handler := new(mockHandler)
	ctx := proxy.NewTestOnlyContext(handler.Handle)
	limiter := &Handler{rate.NewLimiter(1, 0)}

	// WHEN
	limiter.Handle(ctx)

	// THEN
	if assert.NotNil(t, ctx.Response) {
		assert.Equal(t, http.StatusTooManyRequests, ctx.Response.StatusCode)
	}
	handler.AssertNotCalled(t, "Handle", mock.Anything)
}

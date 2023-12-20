/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package limiter

import (
	"github.com/pmateusz/glove/pkg/proxy"
	"golang.org/x/time/rate"
	"net/http"
	"time"
)

type Handler struct {
	*rate.Limiter
}

func (h *Handler) Handle(c *proxy.Context) {
	reservation := h.Reserve()
	if !reservation.OK() {
		c.Response = &http.Response{StatusCode: http.StatusTooManyRequests}
		reservation.Cancel()
		return
	}

	delay := reservation.Delay()
	if delay > 0 {
		time.Sleep(delay)
	}

	c.Next()
}

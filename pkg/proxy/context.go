/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy

import (
	"fmt"
	"github.com/rs/zerolog"
	"net/http"
)

type Context struct {
	Request  *http.Request
	Response *http.Response

	s *session
}

func NewTestOnlyContext(handler Handler) *Context {
	s := &session{
		logger: zerolog.Nop(),
		rule: &Rule{
			Handlers: []Handler{handler},
		},
	}
	return &Context{s: s}
}

func (c *Context) Next() {
	defer func() {
		if r := recover(); r != nil {
			c.s.logger.Error().Str("panic", fmt.Sprintf("%v", r)).Msg("recovered")
			c.s.close = true
			c.Response = newHTTP11Response(http.StatusInternalServerError, c.Request)
		}
	}()

	h := c.s.nextHandler()
	h(c)
}

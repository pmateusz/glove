/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy_test

import (
	"github.com/pmateusz/glove/pkg/proxy"
	"net/http"
	"sync"
)

type engineRunner struct {
	e  *proxy.Engine
	wg sync.WaitGroup
}

func newEngineRunner(options ...proxy.EngineOption) *engineRunner {
	return &engineRunner{
		e: proxy.NewEngine(options...),
	}
}

func (e *engineRunner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.wg.Add(1)
	go func() {
		e.e.ServeHTTP(w, r)
		e.wg.Done()
	}()
}

func (e *engineRunner) Close() {
	e.wg.Wait()
}

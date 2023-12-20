/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package cmd

import "github.com/pmateusz/glove/pkg/proxy"

func Configure(opts ...proxy.EngineOption) {
	engineOptions = append(engineOptions, opts...)
}

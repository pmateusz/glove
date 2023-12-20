/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy

import "crypto/tls"

type Rule struct {
	Action       Action
	ClientConfig func(host string) (*tls.Config, error)
	ServerConfig func(host string) (*tls.Config, error)
	Handlers     []Handler
}

/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy

import (
	"crypto/tls"
	"github.com/rs/zerolog"
	"net"
)

type EngineOptions struct {
	logger *zerolog.Logger
	dialer *net.Dialer

	defaultRule *Rule
	ruleByHost  map[string]*Rule

	clientConfig func(host string) (*tls.Config, error)
	serverConfig func(host string) (*tls.Config, error)
}

func NewEngineOptions() *EngineOptions {
	return &EngineOptions{
		ruleByHost: make(map[string]*Rule),
	}
}

type EngineOption func(ext *EngineOptions)

func WithLogger(logger zerolog.Logger) EngineOption {
	return func(opts *EngineOptions) {
		opts.logger = &logger
	}
}

func WithRule(rule *Rule, host string, otherHosts ...string) EngineOption {
	return func(opts *EngineOptions) {
		opts.ruleByHost[host] = rule
		for _, otherHost := range otherHosts {
			opts.ruleByHost[otherHost] = rule
		}
	}
}

func WithDefaultRule(defaultRule *Rule) EngineOption {
	return func(opts *EngineOptions) {
		opts.defaultRule = defaultRule
	}
}

func WithDialer(dialer *net.Dialer) EngineOption {
	return func(opts *EngineOptions) {
		opts.dialer = dialer
	}
}

func WithClientConfig(clientConfig func(string) (*tls.Config, error)) EngineOption {
	return func(opts *EngineOptions) {
		opts.clientConfig = clientConfig
	}
}

func WithServerConfig(serverConfig func(string) (*tls.Config, error)) EngineOption {
	return func(opts *EngineOptions) {
		opts.serverConfig = serverConfig
	}
}

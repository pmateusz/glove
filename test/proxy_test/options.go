/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/pmateusz/glove/pkg/proxy"
	"net"
	"net/http/httptest"
)

func WithTestServer(server *httptest.Server, mitm bool) proxy.EngineOption {
	return func(opts *proxy.EngineOptions) {
		if cert := server.Certificate(); cert != nil {
			rootCAs := x509.NewCertPool()
			rootCAs.AddCert(cert)
			proxy.WithServerConfig(func(host string) (*tls.Config, error) {
				return &tls.Config{RootCAs: rootCAs}, nil
			})(opts)
		}

		if server.TLS != nil {
			proxy.WithClientConfig(
				func(host string) (*tls.Config, error) {
					return server.TLS, nil
				})(opts)
		}

		if mitm {
			address := server.Listener.Addr().String()
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				panic(fmt.Sprintf("test tools: failed to extract host from address %s: %v", address, err))
			}

			WithMITM(host)(opts)
		}
	}
}

func WithMITM(host string) proxy.EngineOption {
	return proxy.WithRule(&proxy.Rule{Action: proxy.MITMAction}, host)
}

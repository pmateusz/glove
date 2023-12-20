/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy_test

import (
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"github.com/pmateusz/glove/internal/ca"
	"testing"
)

type testCA struct {
	t  *testing.T
	ca *ca.CA
}

func (t *testCA) SignLocalhost() *tls.Certificate {
	serverCert, serverCertErr := t.ca.SignHosts(localhost)
	if serverCertErr != nil {
		t.t.Fatalf("test tools: failed to generate certificate for %s: %v", localhost, serverCertErr)
	}
	return serverCert
}

func (t *testCA) RootCAs() *x509.CertPool {
	certPool := x509.NewCertPool()
	certPool.AddCert(t.ca.Cert())
	return certPool
}

func newCA(t *testing.T) *testCA {
	keyGen := &ca.ECDSAKeyGenerator{Curve: elliptic.P256()}
	ca, caErr := ca.NewCA(keyGen, nil)
	if caErr != nil {
		t.Fatalf("test tools: failed to create certificate authority: %v", caErr)
	}

	return &testCA{
		t,
		ca,
	}
}

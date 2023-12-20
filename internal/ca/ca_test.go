/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package ca

import (
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func serverHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func testGenerateSelfSignedCertificateForTLSServer(t *testing.T, generator KeyGenerator) {
	// GIVEN
	ca, caErr := NewCA(generator, nil)
	require.NoError(t, caErr)
	hostCert, signErr := ca.SignHosts("127.0.0.1")
	require.NoError(t, signErr)
	server := httptest.NewUnstartedServer(http.HandlerFunc(serverHTTP))
	defer server.Close()
	server.TLS = &tls.Config{Certificates: []tls.Certificate{*hostCert}}
	server.StartTLS()

	request, requestErr := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, requestErr)

	certPool := x509.NewCertPool()
	certPool.AddCert(ca.Cert())
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: certPool,
		},
	}

	// WHEN
	resp, roundTripErr := transport.RoundTrip(request)

	// THEN
	assert.NoError(t, roundTripErr)
	if assert.NotNil(t, resp) {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestCreateSelfSignedECDSAForTLSServer(t *testing.T) {
	generator := ECDSAKeyGenerator{Curve: elliptic.P256()}
	testGenerateSelfSignedCertificateForTLSServer(t, &generator)
}

func TestCreateSelfSignedRSAForTLSServer(t *testing.T) {
	generator := RSAKeyGenerator{Bits: 4096}
	testGenerateSelfSignedCertificateForTLSServer(t, &generator)
}

func testCreateSaveAndLoadCA(t *testing.T, generator KeyGenerator) {
	// GIVEN
	ca, createCAErr := NewCA(generator, nil)
	require.NoError(t, createCAErr)
	caPath := t.TempDir() + "/ca.pem"
	privateKeyPath := t.TempDir() + "/privateKey.pem"
	certPEM, encodeCertErr := certificateDERToPEM(ca.Cert().Raw)
	require.NoError(t, encodeCertErr)
	writeCertErr := os.WriteFile(caPath, certPEM, 0644)
	require.NoError(t, writeCertErr)

	privateKeyPEM, encodePrivateKeyErr := privateKeyToPEM(ca.PrivateKey())
	require.NoError(t, encodePrivateKeyErr)
	writePrivateKeyErr := os.WriteFile(privateKeyPath, privateKeyPEM, 0644)
	require.NoError(t, writePrivateKeyErr)

	// WHEN
	loadedCa, loadErr := LoadCA(caPath, privateKeyPath, nil)
	require.NoError(t, loadErr)

	// THEN
	assert.Equal(t, ca.cert, loadedCa.cert)
	assert.Equal(t, ca.privateKey, loadedCa.privateKey)
}

func TestSaveAndRestoreRSACA(t *testing.T) {
	generator := &RSAKeyGenerator{Bits: 4096}
	testCreateSaveAndLoadCA(t, generator)
}

func TestSaveAndRestoreECDSACA(t *testing.T) {
	generator := &ECDSAKeyGenerator{elliptic.P256()}
	testCreateSaveAndLoadCA(t, generator)
}

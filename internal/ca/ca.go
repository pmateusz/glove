/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package ca

import (
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"
)

type CertTemplate func() x509.Certificate

type CA struct {
	cert         *x509.Certificate
	privateKey   crypto.Signer
	generator    KeyGenerator
	certTemplate CertTemplate
}

func (c *CA) Cert() *x509.Certificate {
	return c.cert
}

func (c *CA) PrivateKey() crypto.PrivateKey {
	return c.privateKey
}

var ErrPEMBlockDecode = errors.New("ca: failed to decode PEM block")

func LoadCA(certFile, keyFile string, certTemplate CertTemplate) (*CA, error) {
	certPEMBytes, readCertErr := os.ReadFile(certFile)
	if readCertErr != nil {
		return nil, readCertErr
	}

	certPEMBlock, _ := pem.Decode(certPEMBytes)
	if certPEMBlock == nil {
		return nil, ErrPEMBlockDecode
	}

	cert, parseCertErr := x509.ParseCertificate(certPEMBlock.Bytes)
	if parseCertErr != nil {
		return nil, parseCertErr
	}

	privateKeyPEMBytes, readPrivateKeyErr := os.ReadFile(keyFile)
	if readPrivateKeyErr != nil {
		return nil, readPrivateKeyErr
	}
	privateKeyBlock, _ := pem.Decode(privateKeyPEMBytes)
	if privateKeyBlock == nil {
		return nil, ErrPEMBlockDecode
	}

	privateKey, privateKeyParseErr := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if privateKeyParseErr != nil {
		return nil, privateKeyParseErr
	}

	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		return nil, onUnsupportedPrivateKeyType(privateKey)
	}

	generator, generatorErr := NewKeyGenerator(privateKey)
	if generatorErr != nil {
		return nil, generatorErr
	}

	certTemplateToUse := certTemplate
	if certTemplateToUse == nil {
		certTemplateToUse = newDefaultCert
	}

	return &CA{
		cert:         cert,
		privateKey:   signer,
		generator:    generator,
		certTemplate: certTemplateToUse}, nil
}

func NewCA(generator KeyGenerator, certTemplate CertTemplate) (*CA, error) {
	privateKey, createPrivateKeyErr := generator.Next()
	if createPrivateKeyErr != nil {
		return nil, createPrivateKeyErr
	}

	certTemplateToUse := certTemplate
	if certTemplateToUse == nil {
		certTemplateToUse = newDefaultCert
	}

	unsignedCert := certTemplateToUse()
	setAttributesForCA(&unsignedCert)
	derBytes, createCertErr := x509.CreateCertificate(rand.Reader, &unsignedCert, &unsignedCert, privateKey.Public(), privateKey)
	if createCertErr != nil {
		return nil, createCertErr
	}
	parsedCert, parseCertErr := x509.ParseCertificate(derBytes)
	if parseCertErr != nil {
		return nil, parseCertErr
	}

	return &CA{
		generator:    generator,
		cert:         parsedCert,
		privateKey:   privateKey,
		certTemplate: certTemplateToUse}, nil
}

func (c *CA) sign(cert *x509.Certificate, publicKey any) ([]byte, error) {
	return x509.CreateCertificate(rand.Reader, cert, c.cert, publicKey, c.privateKey)
}

func (c *CA) SignHosts(hosts ...string) (*tls.Certificate, error) {
	privateKey, privateKeyCreateErr := c.generator.Next()
	if privateKeyCreateErr != nil {
		return nil, privateKeyCreateErr
	}

	unsignedCert := c.certTemplate()
	setAttributesForTLS(&unsignedCert, privateKey)
	setHosts(&unsignedCert, hosts)

	signedCert, signErr := c.sign(&unsignedCert, publicKey(privateKey))
	if signErr != nil {
		return nil, signErr
	}

	parsedCert, parseCertErr := x509.ParseCertificate(signedCert)
	if parseCertErr != nil {
		return nil, parseCertErr
	}

	return &tls.Certificate{
		Certificate: [][]byte{signedCert},
		PrivateKey:  privateKey,
		Leaf:        parsedCert,
	}, nil
}

func newDefaultCert() x509.Certificate {
	now := time.Now()
	year, month, day := now.Date()
	notBefore := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	notAfter := notBefore.AddDate(1, 0, 0)
	return x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			Organization: []string{"Glove HTTP Proxy"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
	}
}

func onUnsupportedPrivateKeyType(privateKey crypto.PrivateKey) error {
	return fmt.Errorf("unsupported private key type: %T", privateKey)
}

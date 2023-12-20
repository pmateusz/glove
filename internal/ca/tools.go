/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package ca

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"slices"
)

func certificateDERToPEM(in []byte) ([]byte, error) {
	var buffer bytes.Buffer
	encodeErr := pem.Encode(&buffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: in,
	})
	if encodeErr != nil {
		return nil, encodeErr
	}
	return buffer.Bytes(), nil
}

func privateKeyToPEM(privateKey crypto.PrivateKey) ([]byte, error) {
	var header string

	switch privateKey.(type) {
	case *rsa.PrivateKey:
		header = "RSA PRIVATE KEY"
	case *ecdsa.PrivateKey:
		header = "EC PRIVATE KEY"
	default:
		return nil, fmt.Errorf("unsupported private key type: %T", privateKey)
	}

	derBytes, derEncodeErr := x509.MarshalPKCS8PrivateKey(privateKey)
	if derEncodeErr != nil {
		return nil, derEncodeErr
	}

	var pemBuffer bytes.Buffer
	pemEncodeErr := pem.Encode(&pemBuffer, &pem.Block{
		Type:  header,
		Bytes: derBytes,
	})

	if pemEncodeErr != nil {
		return nil, pemEncodeErr
	}

	return pemBuffer.Bytes(), nil
}

func publicKey(privateKey crypto.PrivateKey) crypto.PublicKey {
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public()
	default:
		return nil
	}
}

func setAttributesForCA(cert *x509.Certificate) {
	cert.IsCA = true
	cert.BasicConstraintsValid = true
	cert.KeyUsage |= x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign
	setExtKeyUsage(cert, x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth)
}

func setAttributesForTLS(cert *x509.Certificate, privateKey crypto.PrivateKey) {
	cert.BasicConstraintsValid = true

	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature KeyUsage bits set in the x509.Config template
	cert.KeyUsage |= x509.KeyUsageDigitalSignature
	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In the context of TLS this KeyUsage
	// is particular to RSA key exchange and authentication.
	if _, isRSA := privateKey.(*rsa.PrivateKey); isRSA {
		cert.KeyUsage |= x509.KeyUsageKeyEncipherment
	}

	setExtKeyUsage(cert, x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth)
}

func setHosts(cert *x509.Certificate, hosts []string) {
	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
		} else {
			cert.DNSNames = append(cert.DNSNames, host)
		}
	}
}

func setExtKeyUsage(cert *x509.Certificate, keyUsages ...x509.ExtKeyUsage) {
	for _, keyUsage := range keyUsages {
		if !slices.Contains(cert.ExtKeyUsage, keyUsage) {
			cert.ExtKeyUsage = append(cert.ExtKeyUsage, keyUsage)
		}
	}
}

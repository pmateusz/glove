/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package ca

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
)

type KeyGenerator interface {
	Next() (crypto.Signer, error)
}

func NewKeyGenerator(privateKey crypto.PrivateKey) (KeyGenerator, error) {
	switch privateKeyToUse := privateKey.(type) {
	case *rsa.PrivateKey:
		return &RSAKeyGenerator{privateKeyToUse.PublicKey.Size()}, nil
	case *ecdsa.PrivateKey:
		return &ECDSAKeyGenerator{Curve: privateKeyToUse.Curve}, nil
	default:
		return nil, onUnsupportedPrivateKeyType(privateKey)
	}
}

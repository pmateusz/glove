/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package ca

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
)

type RSAKeyGenerator struct {
	Bits int
}

func (g *RSAKeyGenerator) Next() (crypto.Signer, error) {
	return rsa.GenerateKey(rand.Reader, g.Bits)
}

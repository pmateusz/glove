/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package ca

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

type ECDSAKeyGenerator struct {
	Curve elliptic.Curve
}

func (g *ECDSAKeyGenerator) Next() (crypto.Signer, error) {
	return ecdsa.GenerateKey(g.Curve, rand.Reader)
}

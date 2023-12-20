/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package acl

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

var localhost = net.IPv4(127, 0, 0, 1)

func TestCreateEmptyACL(t *testing.T) {
	// GIVEN
	acl := NewWhitelist()

	// WHEN
	allowed := acl.Allowed(localhost)

	// THEN
	assert.False(t, allowed)
}

func TestCreateACLWithIP(t *testing.T) {
	// GIVEN
	acl := NewWhitelist(WithIP(localhost))

	// WHEN
	allowed := acl.Allowed(localhost)

	// THEN
	assert.True(t, allowed)
}

func TestCreateACLWithMask(t *testing.T) {
	// GIVEN
	_, network, parseErr := net.ParseCIDR("127.0.0.1/8")
	require.NoError(t, parseErr)
	acl := NewWhitelist(WithMask(*network))

	// WHEN
	allowLocalhost := acl.Allowed(localhost)
	denyOutsideMask := acl.Allowed(net.IPv4(128, 0, 0, 1))

	// THEN
	assert.True(t, allowLocalhost)
	assert.False(t, denyOutsideMask)
}

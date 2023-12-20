/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package version

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseDefaultVersion(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	r := require.New(t)
	value := "0.0.0"

	// WHEN
	ver, err := ParseVersion(value)

	// THEN
	r.NoError(err)
	a.Equal(0, ver.Major)
	a.Equal(0, ver.Minor)
	a.Equal(0, ver.Patch)
	a.Equal("", ver.Suffix)
	a.Equal(value, ver.String())
}

func TestParseGitTagVersion(t *testing.T) {
	// GIVEN
	a := assert.New(t)
	r := require.New(t)
	value := "0.20.2-186-g21514d8c"

	// WHEN
	ver, err := ParseVersion(value)

	// THEN
	r.NoError(err)
	a.Equal(0, ver.Major)
	a.Equal(20, ver.Minor)
	a.Equal(2, ver.Patch)
	a.Equal("-186-g21514d8c", ver.Suffix)
	a.Equal(value, ver.String())
}

func TestCurrentVersion(t *testing.T) {
	// GIVEN
	a := assert.New(t)

	// WHEN
	ver := CurrentVersion()

	// THEN
	a.Equal("0.0.0", ver.String())
}

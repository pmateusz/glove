/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package runtime

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"testing"
)

func testIsContainer(t *testing.T, input string, expectedStatus bool) {
	// GIVEN
	filePath := t.TempDir() + "sample.tmp"
	writeErr := os.WriteFile(filePath, []byte(input), 0644)
	require.NoError(t, writeErr)

	// WHEN
	actualStatus, checkStatusErr := isContainer(filePath)

	// THEN
	assert.Equal(t, expectedStatus, actualStatus)
	assert.NoError(t, checkStatusErr)
}

func TestDockerIsContainer(t *testing.T) {
	testIsContainer(t, "docker", true)
}

func TestLXCIsContainer(t *testing.T) {
	testIsContainer(t, "lxc", true)
}

func TestUbuntuIsContainer(t *testing.T) {
	testIsContainer(t, "0::/init.scope", false)
}

func TestIsContainerReturnsNoErr(t *testing.T) {
	// WHEN
	_, err := IsContainer()

	// THEN
	assert.NoError(t, err)
}

func TestFileNotExists(t *testing.T) {
	// GIVEN
	filePath := t.TempDir() + "sample.tmp"

	// WHEN
	actualStatus, checkStatusErr := isContainer(filePath)

	// THEN
	assert.False(t, actualStatus)
	assert.NoError(t, checkStatusErr)
}

func TestWriteOnlyFile(t *testing.T) {
	// GIVEN
	filePath := t.TempDir() + "sample.tmp"
	writeErr := os.WriteFile(filePath, []byte("test"), 0200)
	require.NoError(t, writeErr)

	// WHEN
	actualStatus, checkStatusErr := isContainer(filePath)

	// THEN
	assert.False(t, actualStatus)
	var pathErr *fs.PathError
	if assert.ErrorAs(t, checkStatusErr, &pathErr) {
		assert.ErrorIs(t, pathErr.Err, fs.ErrPermission)
	}
}

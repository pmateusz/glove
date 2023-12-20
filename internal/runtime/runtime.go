/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package runtime

import (
	"bytes"
	"errors"
	"os"
)

const (
	initProcessCGroupPath = "/proc/1/cgroup"
)

func IsContainer() (bool, error) {
	return isContainer(initProcessCGroupPath)
}

func isContainer(filePath string) (bool, error) {
	buffer := make([]byte, 8)
	nBytes, readErr := readPrefix(filePath, buffer)
	if readErr != nil {
		return false, readErr
	}

	if nBytes == 0 {
		return false, nil
	}

	return bytes.HasPrefix(buffer, []byte("docker")) || bytes.HasPrefix(buffer, []byte("lxc")), nil
}

func readPrefix(filePath string, buffer []byte) (int, error) {
	file, openErr := os.Open(filePath)
	if openErr != nil {
		if errors.Is(openErr, os.ErrNotExist) {
			return 0, nil
		}
		return 0, openErr
	}

	defer func() { _ = file.Close() }()

	return file.Read(buffer)
}

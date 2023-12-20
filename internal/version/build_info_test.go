/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package version

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
	"time"
)

func TestCreateBuildInfo(t *testing.T) {
	// GIVEN
	environment := "local"
	buildTime := "2023-10-28T19:03:34+00:00"
	commitHash := "e28e15b9b3fb93ea55e8af70cbd1a3a0a713e470"
	branch := "master"
	version := "0.0.0"
	a := assert.New(t)

	// WHEN
	buildInfo := newBuildInfo(environment, buildTime, commitHash, branch, version)

	// THEN
	a.Equal(commitHash, buildInfo.CommitHash)
	a.Equal(branch, buildInfo.Branch)
	a.Equal(environment, buildInfo.Environment)
	a.Equal(time.Date(2023, time.October, 28, 19, 3, 34, 0, time.UTC), buildInfo.BuildTime)
	a.Equal(version, buildInfo.Version.String())
	a.Equal(runtime.Version(), buildInfo.GoVersion)
	a.Equal("", buildInfo.GoArch)
	a.Equal("", buildInfo.GoOS)
}

func TestCurrentBuildInfo(t *testing.T) {
	// GIVEN
	a := assert.New(t)

	// WHEN
	buildInfo := CurrentBuildInfo()

	// THEN
	a.Equal("", buildInfo.CommitHash)
	a.Equal("", buildInfo.Branch)
	a.Equal("", buildInfo.Environment)
	a.True(buildInfo.BuildTime.IsZero())
	a.Equal("0.0.0", buildInfo.Version.String())
	a.Equal(runtime.Version(), buildInfo.GoVersion)
	a.Equal("", buildInfo.GoArch) // go1.21.1 linux/amd64 does not fill vcs values
	a.Equal("", buildInfo.GoOS)   // go1.21.1 linux/amd64 does not fill vcs values
}

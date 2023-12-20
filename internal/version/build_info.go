/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package version

import (
	"runtime/debug"
	"time"
)

type BuildInfo struct {
	Version     Version
	BuildTime   time.Time
	Environment string
	Branch      string
	CommitHash  string

	GoVersion string
	GoOS      string
	GoArch    string
}

func newBuildInfo(environment, buildTime, commitHash, branch, version string) BuildInfo {
	parsedVersion, _ := ParseVersion(version)
	parsedTime, _ := time.ParseInLocation(time.RFC3339, buildTime, time.UTC)

	var goVersion, goOs, goArch string
	debugInfo, ok := debug.ReadBuildInfo()
	if ok {
		goVersion = debugInfo.GoVersion
		for _, setting := range debugInfo.Settings {
			switch setting.Key {
			case "GOOS":
				goOs = setting.Value
			case "GOARCH":
				goArch = setting.Value
			}
		}
	}

	return BuildInfo{
		BuildTime:   parsedTime,
		Version:     parsedVersion,
		CommitHash:  commitHash,
		Branch:      branch,
		Environment: environment,
		GoVersion:   goVersion,
		GoOS:        goOs,
		GoArch:      goArch,
	}
}

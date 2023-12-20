/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package version

// version stores the version of the project build in the format "<MAJOR>.<MINOR>.<PATCH><SUFFIX>".
// MAJOR, MINOR AND PATCH are integers, and SUFFIX is an arbitrary string. An empty string is also accepted.
// Example: 0.20.2-186-g21514d8c
// If version format is invalid, the value will be replaced at runtime by `0.0.0`.
var version string

// branch points to the branch of the project VSC that was built.
var branch string

// commitHash indicates the latest commit of the project VSC included in the build.
var commitHash string

// buildTime is datetime when the build process was started. The value should be saved in ISO8604/RFC3339 format.
// UTC time zone is assumed for parsing.
// Example: 2023-10-28T19:03:34+00:00.
// If buildTime format is invalid, it will be replaced at runtime by `zero` time (i.e., `time.Time{}`).
var buildTime string

// environment contains the name of the location or system that executed the build process. It allows to distinguish
// between local and CI builds.
var environment string

func CurrentVersion() Version {
	ver, _ := ParseVersion(version)
	return ver
}

func CurrentBuildInfo() BuildInfo {
	return newBuildInfo(environment, buildTime, commitHash, branch, version)
}

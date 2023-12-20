/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package cmd

import (
	"github.com/pmateusz/glove/internal/version"
	"github.com/spf13/cobra"
	"text/template"
	"time"
)

const versionTemplate = `App:
  Version:     {{.Version}}
Build Info:
  Time:        {{.Build.Time}}
  Commit:      {{.Build.CommitHash}}
  Branch:      {{.Build.Branch}}
  Environment: {{.Build.Environment}}
  GO:
    Version:   {{.Build.GoVersion}}
    Arch:      {{.Build.GoArch}}
    OS Family: {{.Build.GoOS}}
`

type buildValues struct {
	Time        time.Time
	Environment string
	Branch      string
	CommitHash  string
	GoVersion   string
	GoArch      string
	GoOS        string
}

type versionValues struct {
	Version string
	Build   buildValues
}

func newVersionValues(buildInfo version.BuildInfo) versionValues {
	return versionValues{
		Version: buildInfo.Version.String(),
		Build: buildValues{
			Environment: buildInfo.Environment,
			Time:        buildInfo.BuildTime,
			Branch:      buildInfo.Branch,
			CommitHash:  buildInfo.CommitHash,
			GoVersion:   buildInfo.GoVersion,
			GoArch:      buildInfo.GoArch,
			GoOS:        buildInfo.GoOS,
		},
	}
}

func runVersion(cmd *cobra.Command, _ []string) error {
	temp := template.Must(template.New("version").Parse(versionTemplate))
	buildInfo := version.CurrentBuildInfo()
	values := newVersionValues(buildInfo)
	return temp.Execute(cmd.OutOrStdout(), values)
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print build information",
		RunE:  runVersion,
	}
}

/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package cmd

import (
	"github.com/pmateusz/glove/internal/version"
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "glove",
		Short:   "Glove CLI",
		Version: version.CurrentVersion().String(),
	}

	flags := command.PersistentFlags()
	flags.StringVarP(&logMode, "logMode", "", "auto", "set logging mode [auto, console, struct]")
	flags.StringVarP(&logLevel, "logLevel", "", "info", "set logging level [trace, debug, info, warn, error]")

	command.AddCommand(
		newListedCommand(),
		newVersionCommand())
	return command
}

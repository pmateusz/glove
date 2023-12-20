/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package cmd

import "github.com/spf13/cobra"

func Execute() {
	command := newRootCommand()
	err := command.Execute()
	cobra.CheckErr(err)
}

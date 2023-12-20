/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package cmd

import (
	"fmt"
	"github.com/pmateusz/glove/internal/logging"
	"github.com/pmateusz/glove/internal/runtime"
	"github.com/rs/zerolog"
	"strings"
)

var logMode string
var logLevel string

type logSettings struct {
	mode  logging.Mode
	level zerolog.Level
}

func newLoggingSettings(modeName, levelName string) (*logSettings, error) {
	mode, parseModeErr := parseLogMode(modeName)
	if parseModeErr != nil {
		return nil, parseModeErr
	}

	level, parseLevelErr := parseLogLevel(levelName)
	if parseLevelErr != nil {
		return nil, parseLevelErr
	}

	return &logSettings{mode, level}, nil
}

func parseLogLevel(levelName string) (zerolog.Level, error) {
	level, parseErr := zerolog.ParseLevel(logLevel)
	if parseErr != nil {
		return zerolog.NoLevel, fmt.Errorf("failed to parse log level %q, allowed values: \"trace\", \"debug\", \"info\", \"warn\" or \"error\"", levelName)
	}
	return level, nil
}

func parseLogMode(modeName string) (logging.Mode, error) {
	modeNameToUse := strings.ToLower(modeName)
	if modeNameToUse == "" || modeNameToUse == "auto" {
		isContainer, _ := runtime.IsContainer()
		if isContainer {
			return logging.StructMode, nil
		}
		return logging.ConsoleMode, nil
	}

	mode := logging.ParseMode(modeNameToUse)
	if mode == logging.UnknownMode {
		return mode,
			fmt.Errorf("failed to parse log mode %q, allowed values: \"console\" or \"struct\"", mode)
	}

	return logging.UnknownMode, nil
}

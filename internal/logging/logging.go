/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	stdlog "log"
	"os"
	"strings"
	"time"
)

type Mode int8

const (
	UnknownMode Mode = iota
	ConsoleMode
	StructMode
)

func SetupGlobal(mode Mode, level zerolog.Level) {
	switch mode {
	case StructMode:
		SetupStructGlobal(level)
	default:
		SetupConsoleGlobal(level)
	}
}

func SetupStructGlobal(level zerolog.Level) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := NewStructLogger()
	setupGlobal(level, logger)
	return logger
}

func NewStructLogger() zerolog.Logger {
	return zerolog.New(os.Stderr).With().Timestamp().Logger()
}

func SetupConsoleGlobal(level zerolog.Level) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	logger := NewConsoleLogger()
	setupGlobal(level, logger)
	return logger
}

func NewConsoleLogger() zerolog.Logger {
	writer := zerolog.ConsoleWriter{Out: os.Stderr,
		NoColor:    false,
		TimeFormat: zerolog.TimeFieldFormat}
	return zerolog.New(writer).With().Caller().Timestamp().Logger()
}

func setupGlobal(level zerolog.Level, logger zerolog.Logger) {
	zerolog.SetGlobalLevel(level)
	log.Logger = logger
	stdlog.SetFlags(0)
	stdlog.SetOutput(logger)
}

func ParseMode(mode string) Mode {
	switch strings.ToLower(mode) {
	case "console":
		return ConsoleMode
	case "struct":
		return StructMode
	default:
		return UnknownMode
	}
}

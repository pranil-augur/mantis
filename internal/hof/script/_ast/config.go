/*
/*
 * Copyright (c) 2024 Augur AI, Inc.
 * This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. 
 * If a copy of the MPL was not distributed with this file, you can obtain one at https://mozilla.org/MPL/2.0/.
 *
 
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package ast

import (
	"os"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	// Logger
	Logger   *zap.SugaredLogger
	LogLevel string
	logLevel zapcore.Level

	// Filesystem
	FS billy.Filesystem
}

func (P *Parser) initLogger() {
	level := zapcore.ErrorLevel
	if P.config.LogLevel != "" {
		switch strings.ToLower(P.config.LogLevel) {
		case "debug":
			level = zapcore.DebugLevel
		case "info":
			level = zapcore.InfoLevel
		case "warn":
			level = zapcore.WarnLevel
		case "error":
			level = zapcore.ErrorLevel
		case "fatal":
			level = zapcore.FatalLevel
		default:
			panic("invalid log level: " + P.config.LogLevel)
		}
	}

	P.config.logLevel = level

	// filter below a level
	filter := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= P.config.logLevel
	})

	// onsole output
	console := zapcore.Lock(os.Stdout)

	// setup our config and console encoder
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("15:04:05.000"))
	}

	consoleEncoder := zapcore.NewConsoleEncoder(config)

	core := zapcore.NewCore(consoleEncoder, console, filter)
	logger := zap.New(core)
	P.logger = logger.Sugar()
}

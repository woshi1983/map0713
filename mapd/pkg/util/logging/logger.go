// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil/buffer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// TraceLevel logs are typically voluminous, and must be disabled in
// production.
const TraceLevel zapcore.Level = -2

// LogLevel is a wrapper of zapcore.Level, which implements pflag.Value.
type LogLevel struct {
	zapcore.Level
}

func (val LogLevel) String() string {
	if val.Level == TraceLevel {
		return "\"trace\""
	}
	return strconv.Quote(val.Level.String())
}

// Set implements the Set method of pflag.Value.
func (val *LogLevel) Set(str string) error {
	switch str {
	case "trace", "TRACE":
		val.Level = TraceLevel
		return nil
	}
	return val.Level.Set(str)
}

// Type implements the Type method of pflag.Value.
func (val *LogLevel) Type() string {
	return "enum"
}

// NewLogLevelValue assigns a new value to given pointer, and returns the
// updated pointer.
func NewLogLevelValue(val LogLevel, p *LogLevel) *LogLevel {
	*p = val
	return (*LogLevel)(p)
}

// DefaultLogLevelValue is a shorthand of NewLogLevelValue. It sets a default
// value to the given pointer.
func DefaultLogLevelValue(ptr *LogLevel) *LogLevel {
	return NewLogLevelValue(LogLevel{zapcore.InfoLevel}, ptr)
}

// LoggingDispatcher is a global message dispatcher to distribute logging
// messages.
var LoggingDispatcher = buffer.NewDispatcher(1000)

// NewLogger initializes a new root logger with given minimum level.
func NewLogger(level LogLevel, rotateSizeMiB int64, file string) (logger *zap.Logger, err error) {
	var opts []zap.Option
	if level.Level < zapcore.DebugLevel {
		opts = append(opts, zap.Development(), zap.AddCaller())
		level.Level = zapcore.DebugLevel
	}
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	enabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= level.Level
	})
	consoleEncoderCfg := encoderCfg
	if IsColorfulTerm() {
		consoleEncoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var consoleSink zapcore.WriteSyncer
	consoleSink, _, err = zap.Open("stderr")
	if err != nil {
		return
	}
	sink := zap.CombineWriteSyncers(consoleSink, zapcore.AddSync(LoggingDispatcher))
	if file != "" {
		fileSink := zapcore.AddSync(&lumberjack.Logger{
			Filename:   file,
			MaxSize:    int(rotateSizeMiB),
			MaxAge:     30,
			MaxBackups: 3,
			LocalTime:  true,
			Compress:   true,
		})
		sink = zap.CombineWriteSyncers(sink, fileSink)
	}

	core := zapcore.NewCore(zapcore.NewConsoleEncoder(consoleEncoderCfg), sink, enabler)
	return zap.New(core, opts...), nil
}

func IsColorfulTerm() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		t := os.Getenv("TERM")
		if t != "xterm" && strings.HasPrefix(t, "xterm") {
			return true
		}
	}
	return false
}

// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build go1.9

package log

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/util"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultLogLevel = "DEBUG"
)

var (
	StdoutSyncer = zapcore.AddSync(os.Stdout)
	StderrSyncer = zapcore.AddSync(os.Stderr)

	zapLevelMap = map[string]zapcore.Level{
		"DEBUG": zap.DebugLevel,
		"INFO":  zap.InfoLevel,
		"WARN":  zap.WarnLevel,
		"ERROR": zap.ErrorLevel,
		"FATAL": zap.FatalLevel,
	}
)

// Config struct for lager and rotate parameters
type Config struct {
	LoggerLevel string
	LoggerFile  string
	// if false, log print with JSON format
	LogFormatText bool
	// M Bytes
	LogRotateSize  int
	LogBackupCount int
	// days
	LogBackupAge int
	CallerSkip   int
}

func (cfg Config) WithCallerSkip(s int) Config {
	cfg.CallerSkip = s
	return cfg
}

func (cfg Config) WithFile(path string) Config {
	cfg.LoggerFile = path
	return cfg
}

func Configure() Config {
	return Config{
		LoggerLevel:   defaultLogLevel,
		LogFormatText: true,
		CallerSkip:    globalCallerSkip,
	}
}

func toZapConfig(c Config) zapcore.Core {
	// level config
	l, ok := zapLevelMap[strings.ToUpper(c.LoggerLevel)]
	if !ok {
		l = zap.DebugLevel
	}
	var levelEnabler zap.LevelEnablerFunc = func(level zapcore.Level) bool {
		return level >= l
	}

	// log format
	format := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	var enc zapcore.Encoder
	if c.LogFormatText {
		enc = zapcore.NewConsoleEncoder(format)
	} else {
		enc = zapcore.NewJSONEncoder(format)
	}

	// log rotate
	var syncer zapcore.WriteSyncer
	if len(c.LoggerFile) > 0 {
		syncer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   c.LoggerFile,
			MaxSize:    c.LogRotateSize,
			MaxBackups: c.LogBackupCount,
			MaxAge:     c.LogBackupAge,
			LocalTime:  true,
			Compress:   true,
		})
	} else {
		syncer = StdoutSyncer
	}

	zap.NewDevelopment()
	return zapcore.NewCore(enc, syncer, levelEnabler)
}

type Logger struct {
	Config Config

	zapLogger *zap.Logger
	zapSugar  *zap.SugaredLogger
}

func (l *Logger) Debug(msg string) {
	l.zapLogger.Debug(msg)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.zapSugar.Debugf(format, args...)
}

func (l *Logger) Info(msg string) {
	l.zapLogger.Info(msg)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.zapSugar.Infof(format, args...)
}

func (l *Logger) Warn(msg string) {
	l.zapLogger.Warn(msg)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.zapSugar.Warnf(format, args...)
}

func (l *Logger) Error(msg string, err error) {
	if err == nil {
		l.zapLogger.Error(msg)
		return
	}
	l.zapLogger.Error(msg, zap.String("error", err.Error()))
}

func (l *Logger) Errorf(err error, format string, args ...interface{}) {
	if err == nil {
		l.zapSugar.Errorf(format, args...)
		return
	}
	l.zapSugar.With("error", err.Error()).Errorf(format, args...)
}

func (l *Logger) Fatal(msg string, err error) {
	if err == nil {
		l.zapLogger.Panic(msg)
		return
	}
	l.zapLogger.Panic(msg, zap.String("error", err.Error()))
}

func (l *Logger) Fatalf(err error, format string, args ...interface{}) {
	if err == nil {
		l.zapSugar.Panicf(format, args...)
		return
	}
	l.zapSugar.With("error", err.Error()).Panicf(format, args...)
}

// callSkip equals to 0 identify the caller of Recover()
func (l *Logger) Recover(r interface{}, callerSkip int) {
	e := zapcore.Entry{
		Level:  zap.PanicLevel, // zapcore sync automatically when larger than ErrorLevel
		Time:   time.Now(),
		Caller: zapcore.NewEntryCaller(runtime.Caller(callerSkip + 1)),
		Stack:  zap.Stack("stack").String,
	}
	// recover logs also output to stderr
	fmt.Fprintf(StderrSyncer, "%s\tPANIC\t%s\t%s\n%v\n",
		e.Time.Format("2006-01-02T15:04:05.000Z0700"),
		e.Caller.TrimmedPath(),
		r,
		e.Stack)
	if err := l.zapLogger.Core().With([]zap.Field{zap.Reflect("recover", r)}).Write(e, nil); err != nil {
		fmt.Fprintf(StderrSyncer, "%s\tERROR\t%v\n", time.Now().Format("2006-01-02T15:04:05.000Z0700"), err)
		fmt.Fprintln(StderrSyncer, util.BytesToStringWithNoCopy(debug.Stack()))
		StderrSyncer.Sync()
		return
	}
}

func (l *Logger) Sync() {
	l.zapLogger.Sync()
}

func NewLogger(cfg Config) *Logger {
	l := zap.New(toZapConfig(cfg),
		zap.ErrorOutput(StderrSyncer),
		zap.AddCaller(),
		zap.AddCallerSkip(cfg.CallerSkip),
	)
	return &Logger{
		Config:    cfg,
		zapLogger: l,
		zapSugar:  l.Sugar(),
	}
}

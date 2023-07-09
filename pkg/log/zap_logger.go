/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package log

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/openlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	StdoutSyncer = zapcore.Lock(os.Stdout)
	StderrSyncer = zapcore.Lock(os.Stderr)

	zapLevelMap = map[string]zapcore.Level{
		"DEBUG": zap.DebugLevel,
		"INFO":  zap.InfoLevel,
		"WARN":  zap.WarnLevel,
		"ERROR": zap.ErrorLevel,
		"FATAL": zap.FatalLevel,
	}
)

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
	if c.NoCaller {
		format.CallerKey = ""
	}
	if c.NoLevel {
		format.LevelKey = ""
		levelEnabler = func(_ zapcore.Level) bool { return true }
	}
	if c.NoTime {
		format.TimeKey = ""
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

	//zap.NewDevelopment()
	return zapcore.NewCore(enc, syncer, levelEnabler)
}

type ZapLogger struct {
	Config Config

	zapLogger *zap.Logger
	zapSugar  *zap.SugaredLogger
}

func (l *ZapLogger) Debug(msg string, opts ...openlog.Option) {
	l.zapLogger.Debug(msg)
}

func (l *ZapLogger) Info(msg string, opts ...openlog.Option) {
	l.zapLogger.Info(msg)
}

func (l *ZapLogger) Warn(msg string, opts ...openlog.Option) {
	l.zapLogger.Warn(msg)
}

func (l *ZapLogger) Error(msg string, opts ...openlog.Option) {
	options := openlog.ToOptions(opts...)
	if options.Err == nil {
		l.zapLogger.Error(msg)
		return
	}
	l.zapLogger.Error(msg, zap.String("error", options.Err.Error()))
}

func (l *ZapLogger) Fatal(msg string, opts ...openlog.Option) {
	options := openlog.ToOptions(opts...)
	if options.Err == nil {
		l.zapLogger.Panic(msg)
		return
	}
	l.zapLogger.Panic(msg, zap.String("error", options.Err.Error()))
}

// Recover callSkip equals to 0 identify the caller of Recover()
func (l *ZapLogger) Recover(r interface{}, callerSkip int) {
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
	_ = StderrSyncer.Sync() // sync immediately, for server may exit abnormally
	if err := l.zapLogger.Core().With([]zap.Field{zap.Reflect("recover", r)}).Write(e, nil); err != nil {
		fmt.Fprintf(StderrSyncer, "%s\tERROR\t%v\n", time.Now().Format("2006-01-02T15:04:05.000Z0700"), err)
		fmt.Fprintln(StderrSyncer, util.BytesToStringWithNoCopy(debug.Stack()))
		_ = StderrSyncer.Sync()
		return
	}
}

func (l *ZapLogger) Sync() {
	err := l.zapLogger.Sync()
	if err != nil {
		log.Println(err)
	}
	err = StderrSyncer.Sync()
	if err != nil {
		log.Println(err)
	}
	err = StdoutSyncer.Sync()
	if err != nil {
		log.Println(err)
	}
}

func NewZapLogger(cfg Config) *ZapLogger {
	opts := make([]zap.Option, 1)
	opts[0] = zap.ErrorOutput(StderrSyncer)
	if !cfg.NoCaller {
		opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(cfg.CallerSkip))
	}
	l := zap.New(toZapConfig(cfg), opts...)
	if cfg.ReplaceGlobals {
		_ = zap.ReplaceGlobals(l)
	}
	if cfg.RedirectStdLog {
		_ = zap.RedirectStdLog(l)
	}
	logger := &ZapLogger{
		Config:    cfg,
		zapLogger: l,
		zapSugar:  l.Sugar(),
	}
	return logger
}

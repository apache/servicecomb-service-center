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
	"time"

	"github.com/go-chassis/openlog"
)

const (
	globalCallerSkip        = 2
	globalRecoverCallerSkip = 4
	defaultLogLevel         = "DEBUG"
)

var (
	flushFunc   = func() {}
	recoverFunc = func(r interface{}) {}
	Logger      = NewLogger(DefaultConfig())
)

func Init(cfg Config) {
	logger := NewZapLogger(cfg.
		WithCallerSkip(cfg.CallerSkip + globalCallerSkip).
		WithReplaceGlobals(true).
		WithRedirectStdLog(true))
	flushFunc = logger.Sync
	recoverFunc = func(r interface{}) {
		logger.Recover(r, cfg.CallerSkip+globalRecoverCallerSkip)
	}
	Logger = logger
}

func NewLogger(cfg Config) openlog.Logger {
	return NewZapLogger(cfg.WithCallerSkip(cfg.CallerSkip + globalCallerSkip))
}

func DefaultConfig() Config {
	return Config{
		LoggerLevel:   defaultLogLevel,
		LogFormatText: true,
	}
}

func Debug(msg string) {
	Logger.Debug(msg)
}

func Info(msg string) {
	Logger.Info(msg)
}

func Warn(msg string) {
	Logger.Warn(msg)
}

func Error(msg string, err error) {
	Logger.Error(msg, openlog.WithErr(err))
}

func Fatal(msg string, err error) {
	Logger.Fatal(msg, openlog.WithErr(err))
}

func Flush() {
	flushFunc()
}

func NilOrWarn(start time.Time, message string) {
	cost := time.Since(start)
	if cost < time.Second {
		return
	}
	Logger.Warn(fmt.Sprintf("[%s]%s", cost, message))
}

func DebugOrWarn(start time.Time, message string) {
	cost := time.Since(start)
	if cost < time.Second {
		Logger.Debug(fmt.Sprintf("[%s]%s", cost, message))
		return
	}
	Logger.Warn(fmt.Sprintf("[%s]%s", cost, message))
}

func InfoOrWarn(start time.Time, message string) {
	cost := time.Since(start)
	if cost < time.Second {
		Logger.Info(fmt.Sprintf("[%s]%s", cost, message))
		return
	}
	Logger.Warn(fmt.Sprintf("[%s]%s", cost, message))
}

// Panic is a function can only be called in defer function.
func Panic(r interface{}) {
	recoverFunc(r)
}

// Recover is a function call recover() and print the stack in log
// Please call this function like 'defer log.Recover()' in your code
func Recover() {
	if r := recover(); r != nil {
		Panic(r)
	}
}

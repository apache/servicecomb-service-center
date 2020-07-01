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
)

func Default() *Logger {
	return logger
}

func Debug(msg string) {
	logger.Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(msg string) {
	logger.Info(msg)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warn(msg string) {
	logger.Warn(msg)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Error(msg string, err error) {
	logger.Error(msg, err)
}

func Errorf(err error, format string, args ...interface{}) {
	logger.Errorf(err, format, args...)
}

func Fatal(msg string, err error) {
	logger.Fatal(msg, err)
}

func Fatalf(err error, format string, args ...interface{}) {
	logger.Fatalf(err, format, args...)
}

func Sync() {
	logger.Sync()
}

func LogNilOrWarnf(start time.Time, format string, args ...interface{}) {
	cost := time.Since(start)
	if cost < time.Second {
		return
	}
	logger.Warnf("[%s]%s", cost, fmt.Sprintf(format, args...))
}

func LogDebugOrWarnf(start time.Time, format string, args ...interface{}) {
	cost := time.Now().Sub(start)
	if cost < time.Second {
		logger.Debugf("[%s]%s", cost, fmt.Sprintf(format, args...))
		return
	}
	logger.Warnf("[%s]%s", cost, fmt.Sprintf(format, args...))
}

func LogInfoOrWarnf(start time.Time, format string, args ...interface{}) {
	cost := time.Now().Sub(start)
	if cost < time.Second {
		logger.Infof("[%s]%s", cost, fmt.Sprintf(format, args...))
		return
	}
	logger.Warnf("[%s]%s", cost, fmt.Sprintf(format, args...))
}

// LogPanic is a function can only be called in defer function.
func LogPanic(r interface{}) {
	logger.Recover(r, 3) // LogPanic()<-Recover()<-panic()<-final caller
}

// Recover is a function call recover() and print the stack in log
// Please call this function like 'defer log.Recover()' in your code
func Recover() {
	if r := recover(); r != nil {
		LogPanic(r)
	}
	return
}

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

package remote

import (
	"fmt"
	"runtime"

	"github.com/coreos/pkg/capnslog"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

const grpcCallerSkip = 2

// clientLogger implement from grcplog.LoggerV2s and capnslog.Formatter
type clientLogger struct {
}

func (l *clientLogger) Format(pkg string, level capnslog.LogLevel, depth int, entries ...interface{}) {
	format := l.getCaller(depth+1+grpcCallerSkip) + " " + pkg + " %s"
	switch level {
	case capnslog.NOTICE, capnslog.DEBUG, capnslog.TRACE:
		log.Debug(fmt.Sprintf(format, entries...))
	default:
		log.Error(fmt.Sprintf(format, entries...), nil)
	}
}

func (l *clientLogger) getCaller(depth int) string {
	_, file, line, ok := runtime.Caller(depth + 1)
	if !ok {
		return "???"
	}
	return fmt.Sprintf("%s:%d", util.FileLastName(file), line)
}

func (l *clientLogger) Flush() {
	log.Flush()
}

func (l *clientLogger) Debug(args ...interface{}) {
	log.Debug(fmt.Sprint(args...))
}

func (l *clientLogger) Debugln(args ...interface{}) {
	log.Debug(fmt.Sprint(args...))
}

func (l *clientLogger) Debugf(format string, args ...interface{}) {
	log.Debug(fmt.Sprintf(format, args...))
}

func (l *clientLogger) Info(args ...interface{}) {
	log.Info(fmt.Sprint(args...))
}

func (l *clientLogger) Infoln(args ...interface{}) {
	log.Info(fmt.Sprint(args...))
}

func (l *clientLogger) Infof(format string, args ...interface{}) {
	log.Info(fmt.Sprintf(format, args...))
}

func (l *clientLogger) Warning(args ...interface{}) {
	log.Warn(fmt.Sprint(args...))
}

func (l *clientLogger) Warningln(args ...interface{}) {
	log.Warn(fmt.Sprint(args...))
}

func (l *clientLogger) Warningf(format string, args ...interface{}) {
	log.Warn(fmt.Sprintf(format, args...))
}

func (l *clientLogger) Error(args ...interface{}) {
	log.Error(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Errorln(args ...interface{}) {
	log.Error(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Errorf(format string, args ...interface{}) {
	log.Error(fmt.Sprintf(format, args...), nil)
}

// V reports whether verbosity level l is at least the requested verbose level.
func (l *clientLogger) V(_ int) bool {
	return true
}

func (l *clientLogger) Fatal(args ...interface{}) {
	log.Fatal(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Fatalf(format string, args ...interface{}) {
	log.Fatal(fmt.Sprintf(format, args...), nil)
}

func (l *clientLogger) Fatalln(args ...interface{}) {
	log.Fatal(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Print(args ...interface{}) {
	log.Error(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Printf(format string, args ...interface{}) {
	log.Error(fmt.Sprintf(format, args...), nil)
}

func (l *clientLogger) Println(args ...interface{}) {
	log.Error(fmt.Sprint(args...), nil)
}

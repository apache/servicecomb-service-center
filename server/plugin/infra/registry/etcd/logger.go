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
package etcd

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/coreos/pkg/capnslog"
)

// clientLogger implement from grcplog.LoggerV2s and capnslog.Formatter
type clientLogger struct {
}

func (l *clientLogger) Format(pkg string, level capnslog.LogLevel, depth int, entries ...interface{}) {
	switch level {
	case capnslog.CRITICAL:
		l.Fatalf(pkg+": %s", entries...)
	case capnslog.ERROR:
		l.Errorf(pkg+": %s", entries...)
	case capnslog.WARNING:
		l.Warningf(pkg+": %s", entries...)
	case capnslog.NOTICE, capnslog.INFO:
		l.Infof(pkg+": %s", entries...)
	case capnslog.DEBUG, capnslog.TRACE:
		return
	default:
		return
	}
	l.Flush()
}

func (l *clientLogger) Flush() {
}

func (l *clientLogger) Info(args ...interface{}) {
	util.Logger().Info(fmt.Sprint(args...))
}

func (l *clientLogger) Infoln(args ...interface{}) {
	util.Logger().Info(fmt.Sprint(args...))
}

func (l *clientLogger) Infof(format string, args ...interface{}) {
	util.Logger().Infof(format, args...)
}

func (l *clientLogger) Warning(args ...interface{}) {
	util.Logger().Warn(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Warningln(args ...interface{}) {
	util.Logger().Warn(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Warningf(format string, args ...interface{}) {
	util.Logger().Warnf(nil, format, args...)
}

func (l *clientLogger) Error(args ...interface{}) {
	util.Logger().Error(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Errorln(args ...interface{}) {
	util.Logger().Error(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Errorf(format string, args ...interface{}) {
	util.Logger().Errorf(nil, format, args...)
}

// V reports whether verbosity level l is at least the requested verbose level.
func (l *clientLogger) V(_ int) bool {
	return true
}

func (l *clientLogger) Fatal(args ...interface{}) {
	util.Logger().Fatal(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Fatalf(format string, args ...interface{}) {
	util.Logger().Fatalf(nil, format, args...)
}

func (l *clientLogger) Fatalln(args ...interface{}) {
	util.Logger().Fatal(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Print(args ...interface{}) {
	util.Logger().Error(fmt.Sprint(args...), nil)
}

func (l *clientLogger) Printf(format string, args ...interface{}) {
	util.Logger().Errorf(nil, format, args...)
}

func (l *clientLogger) Println(args ...interface{}) {
	util.Logger().Error(fmt.Sprint(args...), nil)
}

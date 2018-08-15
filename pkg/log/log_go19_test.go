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
	"errors"
	"os"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	l := NewLogger(Configure())
	defer func() {
		r := recover()
		LogPanic(r)
		l.Recover(r, 3)
		defer func() {
			l.Recover(recover(), 3)

			defer func() {
				l.Recover(recover(), 3)

				defer Recover()
				Fatalf(nil, "a")
			}()
			Fatal("a", nil)
		}()
		l.Fatalf(nil, "%s", "b")
	}()

	os.Remove("test.log")
	SetGlobal(Configure().WithCallerSkip(2).WithFile("test.log"))

	Debug("a")
	l.Debug("a")
	Debugf("%s", "b")
	l.Debugf("%s", "b")

	Info("a")
	l.Info("a")
	Infof("%s", "b")
	l.Infof("%s", "b")

	Warn("a")
	l.Warn("a")
	Warnf("%s", "b")
	l.Warnf("%s", "b")

	Error("a", nil)
	l.Error("a", nil)
	Errorf(nil, "%s", "a")
	l.Errorf(nil, "%s", "a")
	Error("a", errors.New("error"))
	l.Error("a", errors.New("error"))
	Errorf(errors.New("error"), "%s", "a")
	l.Errorf(errors.New("error"), "%s", "a")

	Sync()
	l.Sync()

	LogNilOrWarnf(time.Now(), "a")
	LogNilOrWarnf(time.Now().Add(-time.Second), "a")
	LogDebugOrWarnf(time.Now(), "a")
	LogDebugOrWarnf(time.Now().Add(-time.Second), "a")
	LogInfoOrWarnf(time.Now(), "a")
	LogInfoOrWarnf(time.Now().Add(-time.Second), "a")

	l.Fatal("a", nil)
}

func BenchmarkZapLogger(b *testing.B) {
	cfg := Configure().WithFile("test.log")
	l := NewLogger(cfg)
	defer os.Remove("test.log")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infof("error test %s", "x")
	}
	l.Sync()
	// 300000	      6078 ns/op	     176 B/op	       7 allocs/op
	b.ReportAllocs()
}

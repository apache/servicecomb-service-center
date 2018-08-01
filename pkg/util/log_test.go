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
package util

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	InitGlobalLogger(LoggerConfig{
		LoggerLevel: "DEBUG",
	})

	CustomLogger("Not Exist", "testDefaultLOGGER")
	l := Logger()
	if l != globalLogger {
		t.Fatalf("should equal to globalLogger")
	}
	CustomLogger("TestLogger", "testFuncName")
	l = Logger()
	if l == globalLogger || l == nil {
		t.Fatalf("should create a new instance for 'TestLogger'")
	}
	s := Logger()
	if l != s {
		t.Fatalf("should be the same globalLogger")
	}
	CustomLogger("github.com/apache/incubator-servicecomb-service-center/pkg/util", "testPkgPath")
	l = Logger()
	if l == globalLogger || l == nil {
		t.Fatalf("should create a new instance for 'util'")
	}
	// l.Infof("OK")

	LogDebugOrWarnf(time.Now().Add(-time.Second), "x")
	LogNilOrWarnf(time.Now().Add(-time.Second), "x")
	LogInfoOrWarnf(time.Now().Add(-time.Second), "x")

	LogDebugOrWarnf(time.Now(), "x")
	LogNilOrWarnf(time.Now(), "x")
	LogInfoOrWarnf(time.Now(), "x")
}

func BenchmarkLogger(b *testing.B) {
	l := Logger()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Debugf("debug test")
			l.Infof("info test")
		}
	})
	// after:	50000	     20964 ns/op	    1296 B/op	      18 allocs/op
	// before:	50000	     31378 ns/op	    2161 B/op	      30 allocs/op
	b.ReportAllocs()
}

func BenchmarkLoggerCustom(b *testing.B) {
	CustomLogger("BenchmarkLoggerCustom", "bmLogger")
	l := Logger()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Debugf("debug test")
			l.Infof("info test")
		}
	})
	// after:	100000	     21374 ns/op	    1296 B/op	      18 allocs/op
	// before:	50000	     21804 ns/op	    2161 B/op	      30 allocs/op
	b.ReportAllocs()
}

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

// +build !go1.9

package log

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	stlager "github.com/go-chassis/paas-lager"
	"github.com/go-chassis/paas-lager/third_party/forked/cloudfoundry/lager"
	"runtime/debug"
)

var (
	defaultLagerConfig = stlager.DefaultConfig()
	stdOutWriters      = []string{"stdout"}
	fileWriters        = []string{"file"}
)

// Config struct for lager and rotate parameters
type Config struct {
	LoggerLevel   string
	LoggerFile    string
	LogFormatText bool
	// MB
	LogRotateSize  int
	LogBackupCount int
	CallerSkip     int
}

func (c Config) WithCallerSkip(s int) Config {
	c.CallerSkip = s
	return c
}

func (c Config) WithFile(path string) Config {
	c.LoggerFile = path
	return c
}

func Configure() Config {
	return fromLagerConfig(defaultLagerConfig)
}

func fromLagerConfig(c *stlager.Config) Config {
	return Config{
		LoggerLevel:   c.LoggerLevel,
		LoggerFile:    c.LoggerFile,
		LogFormatText: c.LogFormatText,
	}
}

func toLagerConfig(c Config) stlager.Config {
	w := fileWriters
	if len(c.LoggerFile) == 0 {
		w = stdOutWriters
	}
	return stlager.Config{
		Writers:       w,
		LoggerLevel:   c.LoggerLevel,
		LoggerFile:    c.LoggerFile,
		LogFormatText: c.LogFormatText,
	}
}

// newLog new log, unsafe
func NewLogger(cfg Config) *Logger {
	if cfg.CallerSkip == 0 {
		cfg.CallerSkip = globalCallerSkip
	}
	stlager.Init(toLagerConfig(cfg))
	return &Logger{
		Config: cfg,
		Logger: stlager.NewLogger(cfg.LoggerFile),
	}
}

type Logger struct {
	Config Config
	lager.Logger
}

func (l *Logger) Recover(r interface{}, callerSkip int) {
	l.Errorf(nil, "recover from panic, %v", r)
	l.Error(util.BytesToStringWithNoCopy(debug.Stack()), nil)
}

func (l *Logger) Sync() {
}

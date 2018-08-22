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
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/go-chassis/paas-lager/third_party/forked/cloudfoundry/lager"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultLogLevel        = "DEBUG"
	globalCallerSkip       = 2
	defaultLogRotatePeriod = 30 * time.Second
)

//log var
var (
	logger         = NewLogger(Configure().WithCallerSkip(globalCallerSkip))
	globalConfig   Config
	globalLogLevel lager.LogLevel
)

func SetGlobal(cfg Config) {
	// renew the global logger
	if len(cfg.LoggerLevel) == 0 {
		cfg.LoggerLevel = defaultLogLevel
	}
	globalConfig = cfg
	logger = NewLogger(cfg)
	// recreate the deleted log file
	switch strings.ToUpper(cfg.LoggerLevel) {
	case "INFO":
		globalLogLevel = lager.INFO
	case "WARN":
		globalLogLevel = lager.WARN
	case "ERROR":
		globalLogLevel = lager.ERROR
	case "FATAL":
		globalLogLevel = lager.FATAL
	default:
		globalLogLevel = lager.DEBUG
	}

	runLogDirRotate(cfg)

	monitorLogFile()
}

func runLogDirRotate(cfg Config) {
	if len(cfg.LoggerFile) == 0 {
		return
	}
	go func() {
		for {
			<-time.After(defaultLogRotatePeriod)
			LogRotate(filepath.Dir(cfg.LoggerFile), cfg.LogRotateSize, cfg.LogBackupCount)
		}
	}()
}

func monitorLogFile() {
	if len(globalConfig.LoggerFile) == 0 {
		return
	}
	go func() {
		for {
			<-time.After(time.Minute)
			Debug(fmt.Sprintf("Check log file at %s", time.Now()))
			if util.PathExist(globalConfig.LoggerFile) {
				continue
			}

			file, err := os.OpenFile(globalConfig.LoggerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				Errorf(err, "Create log file failed.")
				continue
			}

			// TODO Here will lead to file handle leak
			sink := lager.NewReconfigurableSink(lager.NewWriterSink("file", file, lager.DEBUG), globalLogLevel)
			logger.RegisterSink(sink)
			Errorf(nil, "log file is removed, create again.")
		}
	}()
}

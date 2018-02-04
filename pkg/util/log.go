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
	"fmt"
	"github.com/ServiceComb/paas-lager"
	"github.com/ServiceComb/paas-lager/third_party/forked/cloudfoundry/lager"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

//log var
var (
	LOGGER             lager.Logger
	defaultLagerConfig = stlager.DefaultConfig()
	loggerConfig       LoggerConfig
	logLevel           lager.LogLevel

	loggers     map[string]lager.Logger
	loggerNames map[string]string
	loggersMux  sync.RWMutex

	stdOutWriters = []string{"stdout"}
	fileWriters   = []string{"file"}
)

// LoggerConfig struct for lager and rotate parameters
type LoggerConfig struct {
	LoggerLevel     string
	LoggerFile      string
	LogFormatText   bool
	LogRotatePeriod time.Duration
	LogRotateSize   int
	LogBackupCount  int
}

func init() {
	loggers = make(map[string]lager.Logger, 10)
	loggerNames = make(map[string]string, 10)
	// make LOGGER do not be nil, new a stdout logger
	LOGGER = newLogger(fromLagerConfig(defaultLagerConfig))
}

func fromLagerConfig(c *stlager.Config) LoggerConfig {
	return LoggerConfig{
		LoggerLevel:   c.LoggerLevel,
		LoggerFile:    c.LoggerFile,
		LogFormatText: c.LogFormatText,
	}
}

func toLagerConfig(c LoggerConfig) stlager.Config {
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
func newLogger(cfg LoggerConfig) lager.Logger {
	stlager.Init(toLagerConfig(cfg))
	return stlager.NewLogger(cfg.LoggerFile)
}

func InitGlobalLogger(cfg LoggerConfig) {
	// renew the global logger
	if len(cfg.LoggerLevel) == 0 {
		cfg.LoggerLevel = defaultLagerConfig.LoggerLevel
	}
	loggerConfig = cfg
	LOGGER = newLogger(cfg)
	// log rotate
	RunLogDirRotate(cfg)
	// recreate the deleted log file
	switch strings.ToUpper(cfg.LoggerLevel) {
	case "INFO":
		logLevel = lager.INFO
	case "WARN":
		logLevel = lager.WARN
	case "ERROR":
		logLevel = lager.ERROR
	case "FATAL":
		logLevel = lager.FATAL
	default:
		logLevel = lager.DEBUG
	}
	monitorLogFile()
}

func Logger() lager.Logger {
	if len(loggerNames) == 0 {
		return LOGGER
	}
	funcFullName := getCalleeFuncName()

	for prefix, logFile := range loggerNames {
		if strings.Index(prefix, "/") < 0 {
			// function name
			if prefix != funcFullName[strings.LastIndex(funcFullName, " ")+1:] {
				continue
			}
		} else {
			// package name
			if strings.Index(funcFullName, prefix) < 0 {
				continue
			}
		}
		loggersMux.RLock()
		logger, ok := loggers[logFile]
		loggersMux.RUnlock()
		if ok {
			return logger
		}

		loggersMux.Lock()
		logger, ok = loggers[logFile]
		if !ok {
			cfg := loggerConfig
			if len(cfg.LoggerFile) != 0 {
				cfg.LoggerFile = filepath.Join(filepath.Dir(cfg.LoggerFile), logFile+".log")
			}
			logger = newLogger(cfg)
			loggers[logFile] = logger
			LOGGER.Warnf(nil, "match %s, new logger %s for %s", prefix, logFile, funcFullName)
		}
		loggersMux.Unlock()
		return logger
	}

	return LOGGER
}

func getCalleeFuncName() string {
	fullName := ""
	for i := 2; i <= 4; i++ {
		pc, file, _, ok := runtime.Caller(i)

		if strings.Index(file, "/log.go") > 0 {
			continue
		}

		if ok {
			idx := strings.LastIndex(file, "/src/")
			switch {
			case idx >= 0:
				fullName = file[idx+4:]
			default:
				fullName = file
			}

			if f := runtime.FuncForPC(pc); f != nil {
				fullName += " " + FormatFuncName(f.Name())
			}
		}
		break
	}
	return fullName
}

func CustomLogger(pkgOrFunc, fileName string) {
	loggerNames[pkgOrFunc] = fileName
}

func monitorLogFile() {
	if len(loggerConfig.LoggerFile) == 0 {
		return
	}
	Go(func(stopCh <-chan struct{}) {
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(time.Minute):
				Logger().Debug(fmt.Sprintf("Check log file at %s", time.Now()))

				if !PathExist(loggerConfig.LoggerFile) {
					file, err := os.OpenFile(loggerConfig.LoggerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
					if err != nil {
						Logger().Errorf(err, "Create log file failed.")
						return
					}
					// TODO Here will lead to file handle leak
					sink := lager.NewReconfigurableSink(lager.NewWriterSink("file", file, lager.DEBUG), logLevel)
					Logger().RegisterSink(sink)
					Logger().Errorf(nil, "log file is removed, create again.")
				}
			}
		}
	})
}

func LogNilOrWarnf(start time.Time, format string, args ...interface{}) {
	cost := time.Now().Sub(start)
	if cost < time.Second {
		return
	}
	Logger().Warnf(nil, "[%s]%s", cost, fmt.Sprintf(format, args...))
}

func LogDebugOrWarnf(start time.Time, format string, args ...interface{}) {
	cost := time.Now().Sub(start)
	if cost < time.Second {
		Logger().Debugf("[%s]%s", cost, fmt.Sprintf(format, args...))
		return
	}
	Logger().Warnf(nil, "[%s]%s", cost, fmt.Sprintf(format, args...))
}

func LogInfoOrWarnf(start time.Time, format string, args ...interface{}) {
	cost := time.Now().Sub(start)
	if cost < time.Second {
		Logger().Infof("[%s]%s", cost, fmt.Sprintf(format, args...))
		return
	}
	Logger().Warnf(nil, "[%s]%s", cost, fmt.Sprintf(format, args...))
}

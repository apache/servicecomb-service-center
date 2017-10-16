//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package util

import (
	"fmt"
	"github.com/ServiceComb/service-center/pkg/lager"
	"github.com/ServiceComb/service-center/pkg/lager/core"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

//log var
var (
	LOGGER          core.Logger
	reBuildLogLevel core.LogLevel

	loggers     map[string]core.Logger
	loggerNames map[string]string
	loggersMux  sync.RWMutex
)

func init() {
	loggers = make(map[string]core.Logger, 10)
	loggerNames = make(map[string]string, 10)
	LOGGER = lager.NewLogger("default")
}

func InitLogger(loggerName string, cfg *lager.Config) {
	lager.Init(*cfg)
	LOGGER = lager.NewLogger(loggerName)
	LOGGER.Debug("init logger")

	switch strings.ToUpper(lager.GetConfig().LoggerLevel) {
	case "DEBUG":
		reBuildLogLevel = core.DEBUG
	case "INFO":
		reBuildLogLevel = core.INFO
	case "WARN":
		reBuildLogLevel = core.WARN
	case "ERROR":
		reBuildLogLevel = core.ERROR
	case "FATAL":
		reBuildLogLevel = core.FATAL
	default:
		panic(fmt.Errorf("unknown logger level: %s", lager.GetConfig().LoggerLevel))
	}

	monitorLogFile()
}

func NewLogger(loggerName string, cfg *lager.Config) core.Logger {
	return lager.NewLoggerExt(loggerName, loggerName, cfg)
}

func Logger() core.Logger {
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
			cfg := *lager.GetConfig()
			if len(cfg.LoggerFile) != 0 {
				cfg.LoggerFile = filepath.Join(filepath.Dir(cfg.LoggerFile), logFile+".log")
			}
			logger = NewLogger(logFile, &cfg)
			loggers[logFile] = logger
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

		if strings.Index(file, "log.go") > 0 {
			continue
		}

		if ok {
			idx := strings.LastIndex(file, "src")
			switch {
			case idx >= 0:
				fullName = file[idx+4:]
			default:
				fullName = file
			}

			if f := runtime.FuncForPC(pc); f != nil {
				fullName += " " + formatFuncName(f.Name())
			}
		}
		break
	}
	return fullName
}

func formatFuncName(f string) string {
	i := strings.LastIndex(f, "/")
	j := strings.Index(f[i+1:], ".")
	if j < 1 {
		return "???"
	}
	_, fun := f[:i+j+1], f[i+j+2:]
	i = strings.LastIndex(fun, ".")
	return fun[i+1:]
}

func CustomLogger(pkgOrFunc, fileName string) {
	loggerNames[pkgOrFunc] = fileName
}

func monitorLogFile() {
	Go(func(stopCh <-chan struct{}) {
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(time.Minute):
				Logger().Debug(fmt.Sprintf("Check log file at %s", time.Now()))

				if lager.GetConfig().LoggerFile != "" && !PathExist(lager.GetConfig().LoggerFile) {
					file, err := os.OpenFile(lager.GetConfig().LoggerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
					if err != nil {
						Logger().Errorf(err, "Create log file failed.")
						return
					}
					// TODO Here will lead to file handle leak
					sink := core.NewReconfigurableSink(core.NewWriterSink(file, core.DEBUG), reBuildLogLevel)
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

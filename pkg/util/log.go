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
	"os"

	"bufio"
	"bytes"
	"github.com/ServiceComb/service-center/pkg/lager"
	"github.com/ServiceComb/service-center/pkg/lager/core"
	"path/filepath"
	"runtime/debug"
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
	funcFullName := getCalleeFuncName(debug.Stack())

	for prefix, logFile := range loggerNames {
		if strings.Index(prefix, "/") < 0 {
			// function name
			if prefix != funcFullName[strings.LastIndex(funcFullName, ".")+1:] {
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

func getCalleeFuncName(stack []byte) string {
	reader := bufio.NewReader(bytes.NewReader(stack))
	/*
		goroutine 1 [running]:
		runtime/debug.Stack(0x0, 0x0, 0x0)
			runtime/debug/stack.go:24 +0xbe
		github.com/ServiceComb/service-center/util.Logger(0x0, 0x0)
			github.com/ServiceComb/service-center/util/log.go:67 +0xf2
	*/
	for i := 0; i < 1+2*2; i++ {
		reader.ReadLine()
	}
	line, _, _ := reader.ReadLine()
	funcFullName := BytesToStringWithNoCopy(line)
	funcFullName = funcFullName[:strings.LastIndex(funcFullName, "(")]
	return funcFullName
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

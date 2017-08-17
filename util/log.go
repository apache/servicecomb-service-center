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

	"github.com/ServiceComb/service-center/pkg/lager"
	"github.com/ServiceComb/service-center/pkg/lager/core"
	"strings"
	"time"
)

//log var
var (
	LOGGER        core.Logger
	lagerLogLevel core.LogLevel
)

func InitLogger(loggerName string, cfg *lager.Config) {
	lager.Init(*cfg)
	LOGGER = lager.NewLogger(loggerName)
	LOGGER.Debug("init logger")

	switch strings.ToUpper(lager.GetConfig().LoggerLevel) {
	case "DEBUG":
		lagerLogLevel = core.DEBUG
	case "INFO":
		lagerLogLevel = core.INFO
	case "WARN":
		lagerLogLevel = core.WARN
	case "ERROR":
		lagerLogLevel = core.ERROR
	case "FATAL":
		lagerLogLevel = core.FATAL
	default:
		panic(fmt.Errorf("unknown logger level: %s", lager.GetConfig().LoggerLevel))
	}

	monitorLogFile()
}

func monitorLogFile() {
	Go(func(stopCh <-chan struct{}) {
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(time.Minute):
				LOGGER.Debug(fmt.Sprintf("Check log file at %s", time.Now()))

				if lager.GetConfig().LoggerFile != "" && !PathExist(lager.GetConfig().LoggerFile) {
					file, err := os.OpenFile(lager.GetConfig().LoggerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
					if err != nil {
						LOGGER.Errorf(err, "Create log file failed.")
						return
					}

					sink := core.NewReconfigurableSink(core.NewWriterSink(file, core.DEBUG), lagerLogLevel)
					LOGGER.RegisterSink(sink)
					LOGGER.Errorf(nil, "log file is removed, create again.")
				}
			}
		}
	})
}

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
package lager

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ServiceComb/service-center/pkg/lager/core"
	"github.com/ServiceComb/service-center/pkg/lager/syslog"
)

const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
	FATAL = "FATAL"
)

type Config struct {
	LoggerLevel string
	LoggerFile  string

	EnableRsyslog  bool
	RsyslogNetwork string
	RsyslogAddr    string

	LogFormatText bool

	EnableStdOut bool
}

func GetConfig() *Config {
	return config
}

var config *Config = DefaultConfig()

func DefaultConfig() *Config {
	return &Config{
		LoggerLevel:    INFO,
		LoggerFile:     "",
		EnableRsyslog:  false,
		RsyslogNetwork: "udp",
		RsyslogAddr:    "127.0.0.1:5140",
		LogFormatText:  false,
		EnableStdOut:   false,
	}
}

func Init(c Config) {
	if c.LoggerLevel != "" {
		config.LoggerLevel = c.LoggerLevel
	}

	if c.LoggerFile != "" {
		config.LoggerFile = c.LoggerFile
	}

	if c.EnableRsyslog {
		config.EnableRsyslog = c.EnableRsyslog
	}

	if c.RsyslogNetwork != "" {
		config.RsyslogNetwork = c.RsyslogNetwork
	}

	if c.RsyslogAddr != "" {
		config.RsyslogAddr = c.RsyslogAddr
	}

	config.EnableStdOut = c.EnableStdOut
	config.LogFormatText = c.LogFormatText
}

func NewLogger(component string) core.Logger {
	return NewLoggerExt(component, component)
}

func NewLoggerExt(component string, app_guid string) core.Logger {
	var lagerLogLevel core.LogLevel
	switch strings.ToUpper(config.LoggerLevel) {
	case DEBUG:
		lagerLogLevel = core.DEBUG
	case INFO:
		lagerLogLevel = core.INFO
	case WARN:
		lagerLogLevel = core.WARN
	case ERROR:
		lagerLogLevel = core.ERROR
	case FATAL:
		lagerLogLevel = core.FATAL
	default:
		panic(fmt.Errorf("unknown logger level: %s", config.LoggerLevel))
	}

	logger := core.NewLoggerExt(component, config.LogFormatText)
	if config.EnableStdOut {
		sink := core.NewReconfigurableSink(core.NewWriterSink(os.Stdout, core.DEBUG), lagerLogLevel)
		logger.RegisterSink(sink)
	}
	if config.LoggerFile != "" {
		file, err := os.OpenFile(config.LoggerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}

		sink := core.NewReconfigurableSink(core.NewWriterSink(file, core.DEBUG), lagerLogLevel)
		logger.RegisterSink(sink)
	}

	if config.EnableRsyslog {
		syslog, err := syslog.Dial(component, app_guid, config.RsyslogNetwork, config.RsyslogAddr)
		if err != nil {
			//warn, not panic
			log.Println(err.Error())
		} else {
			sink := core.NewReconfigurableSink(core.NewWriterSink(syslog, core.DEBUG), lagerLogLevel)
			logger.RegisterSink(sink)
		}
	}

	return logger
}

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

package core

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	// import the grace package and parse grace cmd line
	_ "github.com/apache/servicecomb-service-center/pkg/grace"

	"github.com/apache/servicecomb-service-center/pkg/goutil"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/metrics"
)

const (
	defaultCollectPeriod = 30 * time.Second
)

func Initialize() {
	// initialize configuration
	config.Init()
	// Logging
	initLogger()
	// go pool
	goutil.Init()
	// init the sc registration
	InitRegistration()
	// Register global services
	RegisterGlobalServices()
	// init metrics
	initMetrics()
}

func initLogger() {
	log.Init(log.Config{
		LoggerLevel:    config.GetLog().LogLevel,
		LoggerFile:     os.ExpandEnv(config.GetLog().LogFilePath),
		LogFormatText:  config.GetLog().LogFormat == "text",
		LogRotateSize:  int(config.GetLog().LogRotateSize),
		LogBackupCount: int(config.GetLog().LogBackupCount),
	})
}

func initMetrics() {
	if !config.GetBool("metrics.enable", false) {
		return
	}
	interval, err := time.ParseDuration(strings.TrimSpace(config.GetString("metrics.interval", defaultCollectPeriod.String())))
	if err != nil {
		log.Error(fmt.Sprintf("invalid metrics config[interval], set default %s", defaultCollectPeriod), err)
	}
	if interval <= time.Second {
		interval = defaultCollectPeriod
	}
	instance := net.JoinHostPort(config.GetString("server.host", "", config.WithStandby("httpaddr")),
		config.GetString("server.port", "", config.WithStandby("httpport")))

	if err := metrics.Init(metrics.Options{
		Interval: interval,
		Instance: instance,
	}); err != nil {
		log.Fatal("init metrics failed", err)
	}
}

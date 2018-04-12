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
	"flag"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/grace"
	"github.com/apache/incubator-servicecomb-service-center/pkg/plugin"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/version"
	"golang.org/x/net/context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func init() {
	Initialize()
}

func Initialize() {
	initCommandLine()

	plugin.SetPluginDir(ServerInfo.Config.PluginsDir)

	initLogger()

	printVersion()

	util.Go(handleSignals)

	grace.Init()
}

func initCommandLine() {
	var printVer bool
	flag.BoolVar(&printVer, "v", false, "Print the version and exit.")
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.Parse(os.Args[1:])

	if printVer {
		fmt.Printf("ServiceCenter version: %s\n", version.Ver().Version)
		fmt.Printf("Build tag: %s\n", version.Ver().BuildTag)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
}

func printVersion() {
	util.Logger().Infof("service center version: %s", version.Ver().Version)
	util.Logger().Infof("Build tag: %s", version.Ver().BuildTag)
	util.Logger().Infof("Go version: %s", runtime.Version())
	util.Logger().Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)

	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)
	util.Logger().Infof("service center is running simultaneously with %d CPU cores", cores)
}

func initLogger() {
	util.InitGlobalLogger(util.LoggerConfig{
		LoggerLevel:     ServerInfo.Config.LogLevel,
		LoggerFile:      os.ExpandEnv(ServerInfo.Config.LogFilePath),
		LogFormatText:   ServerInfo.Config.LogFormat == "text",
		LogRotatePeriod: 30 * time.Second,
		LogRotateSize:   int(ServerInfo.Config.LogRotateSize),
		LogBackupCount:  int(ServerInfo.Config.LogBackupCount),
	})
}

func handleSignals(ctx context.Context) {
	var sig os.Signal
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM,
	)
	wait := 60 * time.Second
	for {
		select {
		case <-ctx.Done():
			return
		case sig = <-sigCh:
			switch sig {
			case syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM:
				select {
				case <-ctx.Done():
					return
				case <-time.After(wait):
				}
				util.Logger().Warnf(nil, "Waiting for server response timed out(%s), force shutdown.", wait)
				os.Exit(1)
			}
		}
	}
}

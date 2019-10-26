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
	"os"
	"runtime"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/version"
	"github.com/astaxie/beego"
)

const (
	InitVersion = "0"

	defaultServiceClearInterval = 12 * time.Hour //0.5 day
	defaultServiceTTL           = 24 * time.Hour //1 day

	minServiceClearInterval = 30 * time.Second
	minServiceTTL           = 30 * time.Second

	maxServiceClearInterval = 24 * time.Hour       //1 day
	maxServiceTTL           = 24 * 365 * time.Hour //1 year
)

var ServerInfo = pb.NewServerInformation()

func Configure() {
	setCPUs()

	*ServerInfo = newInfo()

	plugin.SetPluginDir(ServerInfo.Config.PluginsDir)

	initLogger()

	version.Ver().Log()
}

func newInfo() pb.ServerInformation {
	maxLogFileSize := beego.AppConfig.DefaultInt64("log_rotate_size", 20)
	if maxLogFileSize <= 0 || maxLogFileSize > 50 {
		maxLogFileSize = 20
	}
	maxLogBackupCount := beego.AppConfig.DefaultInt64("log_backup_count", 50)
	if maxLogBackupCount < 0 || maxLogBackupCount > 100 {
		maxLogBackupCount = 50
	}

	serviceClearInterval, err := time.ParseDuration(os.Getenv("SERVICE_CLEAR_INTERVAL"))
	if err != nil || serviceClearInterval < minServiceClearInterval || serviceClearInterval > maxServiceClearInterval {
		serviceClearInterval = defaultServiceClearInterval
	}

	serviceTTL, err := time.ParseDuration(os.Getenv("SERVICE_TTL"))
	if err != nil || serviceTTL < minServiceTTL || serviceTTL > maxServiceTTL {
		serviceTTL = defaultServiceTTL
	}

	logLevelEnv := os.Getenv("SC_LOG_LEVEL")
	logFilePathEnv := os.Getenv("SC_LOG_FILEPATH")

	logFilePath := beego.AppConfig.String("logfile")
	if logFilePathEnv!= ""{
		logFilePath = logFilePathEnv
	}

	loglevel :=  beego.AppConfig.String("loglevel");
	if logLevelEnv!= ""{
		loglevel = logLevelEnv
	}

	return pb.ServerInformation{
		Version: InitVersion,
		Config: pb.ServerConfig{
			MaxHeaderBytes: int64(beego.AppConfig.DefaultInt("max_header_bytes", 16384)),
			MaxBodyBytes:   beego.AppConfig.DefaultInt64("max_body_bytes", 2097152),

			ReadHeaderTimeout: beego.AppConfig.DefaultString("read_header_timeout", "60s"),
			ReadTimeout:       beego.AppConfig.DefaultString("read_timeout", "60s"),
			IdleTimeout:       beego.AppConfig.DefaultString("idle_timeout", "60s"),
			WriteTimeout:      beego.AppConfig.DefaultString("write_timeout", "60s"),

			LimitTTLUnit:     beego.AppConfig.DefaultString("limit_ttl", "s"),
			LimitConnections: int64(beego.AppConfig.DefaultInt("limit_conns", 0)),
			LimitIPLookup: beego.AppConfig.DefaultString("limit_iplookups",
				"RemoteAddr,X-Forwarded-For,X-Real-IP"),

			SslEnabled:    beego.AppConfig.DefaultInt("ssl_mode", 1) != 0,
			SslMinVersion: beego.AppConfig.DefaultString("ssl_min_version", "TLSv1.2"),
			SslVerifyPeer: beego.AppConfig.DefaultInt("ssl_verify_client", 1) != 0,
			SslCiphers:    beego.AppConfig.String("ssl_ciphers"),

			AutoSyncInterval:  beego.AppConfig.DefaultString("auto_sync_interval", "30s"),
			CompactIndexDelta: beego.AppConfig.DefaultInt64("compact_index_delta", 100),
			CompactInterval:   beego.AppConfig.String("compact_interval"),

			LogRotateSize:  maxLogFileSize,
			LogBackupCount: maxLogBackupCount,
			LogFilePath:    logFilePath,
			LogLevel:       loglevel,
			LogFormat:      beego.AppConfig.DefaultString("log_format", "text"),
			LogSys:         beego.AppConfig.DefaultBool("log_sys", false),

			PluginsDir: beego.AppConfig.DefaultString("plugins_dir", "./plugins"),
			Plugins:    util.NewJSONObject(),

			EnablePProf:  beego.AppConfig.DefaultInt("enable_pprof", 0) != 0,
			EnableCache:  beego.AppConfig.DefaultInt("enable_cache", 1) != 0,
			SelfRegister: beego.AppConfig.DefaultInt("self_register", 1) != 0,

			ServiceClearEnabled:  os.Getenv("SERVICE_CLEAR_ENABLED") == "true",
			ServiceClearInterval: serviceClearInterval,
			ServiceTTL:           serviceTTL,
		},
	}
}

func setCPUs() {
	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)
	log.Infof("service center is running simultaneously with %d CPU cores", cores)
}

func initLogger() {
	log.SetGlobal(log.Config{
		LoggerLevel:    ServerInfo.Config.LogLevel,
		LoggerFile:     os.ExpandEnv(ServerInfo.Config.LogFilePath),
		LogFormatText:  ServerInfo.Config.LogFormat == "text",
		LogRotateSize:  int(ServerInfo.Config.LogRotateSize),
		LogBackupCount: int(ServerInfo.Config.LogBackupCount),
	})
}

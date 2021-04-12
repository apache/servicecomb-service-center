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

package config

import (
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-chassis/go-archaius"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/version"
)

const (
	InitVersion = "0"

	defaultServiceClearInterval = 12 * time.Hour //0.5 day
	defaultServiceTTL           = 24 * time.Hour //1 day

	minServiceClearInterval = 30 * time.Second
	minServiceTTL           = 30 * time.Second
	minCacheTTL             = 5 * time.Minute

	maxServiceClearInterval = 24 * time.Hour       //1 day
	maxServiceTTL           = 24 * 365 * time.Hour //1 year
)

//Configurations is kie config items
var Configurations = &Config{}

var ServerInfo = NewServerInformation()

//GetGov return governance configs
func GetGov() *Gov {
	return Configurations.Gov
}

//GetServer return the http server configs
func GetServer() ServerConfig {
	return Configurations.Server.Config
}

//GetSSL return the ssl configs
func GetSSL() ServerConfig {
	return Configurations.Server.Config
}

//GetLog return the log configs
func GetLog() ServerConfig {
	return Configurations.Server.Config
}

//GetRegistry return the registry configs
func GetRegistry() ServerConfig {
	return Configurations.Server.Config
}

//GetPlugin return the plugin configs
func GetPlugin() ServerConfig {
	return Configurations.Server.Config
}

//GetRBAC return the rbac configs
func GetRBAC() ServerConfig {
	return Configurations.Server.Config
}

//GetMetrics return the metrics configs
func GetMetrics() ServerConfig {
	return Configurations.Server.Config
}

func Init() {
	setCPUs()

	err := archaius.Init(archaius.WithMemorySource(), archaius.WithENVSource(),
		archaius.WithOptionalFiles([]string{filepath.Join(util.GetAppRoot(), "conf", "app.yaml")}))
	if err != nil {
		log.Fatal("can not init archaius", err)
	}
	err = archaius.UnmarshalConfig(Configurations)
	if err != nil {
		log.Fatal("archaius unmarshal config failed", err)
	}
	Configurations.Server = ServerInfo
	*ServerInfo = newInfo()

	plugin.RegisterConfigurator(Configurations)

	version.Ver().Log()
}

func newInfo() ServerInformation {
	serviceClearInterval := GetDuration("registry.service.clearInterval", defaultServiceClearInterval, WithENV("SERVICE_CLEAR_INTERVAL"))
	if serviceClearInterval < minServiceClearInterval || serviceClearInterval > maxServiceClearInterval {
		serviceClearInterval = defaultServiceClearInterval
	}
	serviceTTL := GetDuration("registry.service.clearTTL", defaultServiceTTL, WithENV("SERVICE_TTL"))
	if serviceTTL < minServiceTTL || serviceTTL > maxServiceTTL {
		serviceTTL = defaultServiceTTL
	}
	cacheTTL := GetDuration("registry.cache.ttl", minCacheTTL, WithENV("CACHE_TTL"), WithStandby("cache_ttl"))
	if cacheTTL < minCacheTTL {
		cacheTTL = minCacheTTL
	}

	maxLogFileSize := GetInt64("log.rotateSize", 20, WithStandby("log_rotate_size"))
	if maxLogFileSize <= 0 || maxLogFileSize > 50 {
		maxLogFileSize = 20
	}
	maxLogBackupCount := GetInt64("log.backupCount", 50, WithStandby("log_backup_count"))
	if maxLogBackupCount < 0 || maxLogBackupCount > 100 {
		maxLogBackupCount = 50
	}
	accessLogFile := GetString("log.accessFile", "./access.log", WithENV("SC_ACCESS_LOG_FILE"), WithStandby("access_log_file"))

	return ServerInformation{
		Version: InitVersion,
		Config: ServerConfig{
			MaxHeaderBytes:    GetInt64("server.request.maxHeaderBytes", 16384, WithStandby("max_header_bytes")),
			MaxBodyBytes:      GetInt64("server.request.maxBodyBytes", 2097152, WithStandby("max_body_bytes")),
			ReadHeaderTimeout: GetString("server.request.headerTimeout", "60s", WithStandby("read_header_timeout")),
			ReadTimeout:       GetString("server.request.timeout", "60s", WithStandby("read_timeout")),
			IdleTimeout:       GetString("server.idle.timeout", "60s", WithStandby("idle_timeout")),
			WriteTimeout:      GetString("server.response.timeout", "60s", WithStandby("write_timeout")),

			LimitTTLUnit:     GetString("server.limit.unit", "s", WithStandby("limit_ttl")),
			LimitConnections: GetInt64("server.limit.connections", 0, WithStandby("limit_conns")),
			LimitIPLookup:    GetString("server.limit.ipLookups", "RemoteAddr,X-Forwarded-For,X-Real-IP", WithStandby("limit_iplookups")),

			EnablePProf: GetInt("server.pprof.mode", 0, WithStandby("enable_pprof")) != 0,

			SslEnabled: GetInt("ssl.mode", 1, WithStandby("ssl_mode")) != 0,

			LogRotateSize:   maxLogFileSize,
			LogBackupCount:  maxLogBackupCount,
			LogFilePath:     GetString("log.file", "", WithStandby("logfile")),
			LogLevel:        GetString("log.level", "", WithStandby("loglevel")),
			LogFormat:       GetString("log.format", "text", WithStandby("log_format")),
			LogSys:          GetBool("log.system", false, WithStandby("log_sys")),
			EnableAccessLog: GetBool("log.accessEnable", false, WithStandby("enable_access_log")),
			AccessLogFile:   accessLogFile,

			PluginsDir: GetString("plugin.dir", "./plugins", WithStandby("plugins_dir")),
			Plugins:    util.NewJSONObject(),

			EnableCache:  GetInt("registry.cache.mode", 1, WithStandby("enable_cache")) != 0,
			CacheTTL:     cacheTTL,
			SelfRegister: GetInt("registry.selfRegister", 1, WithStandby("self_register")) != 0,

			ServiceClearEnabled:  GetBool("registry.service.clearEnable", false, WithENV("SERVICE_CLEAR_ENABLED")),
			ServiceClearInterval: serviceClearInterval,
			ServiceTTL:           serviceTTL,
			GlobalVisible:        GetString("registry.service.globalVisible", "", WithENV("CSE_SHARED_SERVICES")),
			InstanceTTL:          GetInt64("registry.instance.ttl", 0, WithENV("INSTANCE_TTL")),

			SchemaDisable:  GetBool("registry.schema.disable", false, WithENV("SCHEMA_DISABLE")),
			SchemaEditable: GetBool("registry.schema.editable", false, WithENV("SCHEMA_EDITABLE")),

			EnableRBAC: GetBool("rbac.enable", false, WithStandby("rbac_enabled")),

			MetricsInterval: GetDuration("metrics.interval", 30*time.Second, WithENV("METRICS_INTERVAL")),
		},
	}
}

func setCPUs() {
	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)
	log.Infof("service center is running simultaneously with %d CPU cores", cores)
}

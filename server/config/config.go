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
	"encoding/json"
	"fmt"
	"os"
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
	minCacheTTL = 60 * time.Minute
)

var (
	Server = NewServerConfig()
	// App is application root config
	App = &AppConfig{Server: Server}
)

// GetProfile return active profile
func GetProfile() *ServerConfig {
	return App.Server
}

// GetGov return governance configs
func GetGov() *Gov {
	return App.Gov
}

// GetServer return the http server configs
func GetServer() ServerConfigDetail {
	return App.Server.Config
}

// GetSSL return the ssl configs
func GetSSL() ServerConfigDetail {
	return App.Server.Config
}

// GetLog return the log configs
func GetLog() ServerConfigDetail {
	return App.Server.Config
}

// GetRegistry return the registry configs
func GetRegistry() ServerConfigDetail {
	return App.Server.Config
}

// GetPlugin return the plugin configs
func GetPlugin() ServerConfigDetail {
	return App.Server.Config
}

// GetRBAC return the rbac configs
func GetRBAC() ServerConfigDetail {
	return App.Server.Config
}

func Init() {
	setCPUs()

	err := archaius.Init(archaius.WithMemorySource(), archaius.WithENVSource())
	if err != nil {
		log.Fatal("can not init archaius", err)
	}

	err = archaius.AddFile(filepath.Join(util.GetAppRoot(), "conf", "app.yaml"))
	if err != nil {
		log.Warn(fmt.Sprintf("can not add app config file source, error: %s", err))
	}

	err = Reload()
	if err != nil {
		log.Fatal("reload configs failed", err)
	}

	plugin.RegisterConfigurator(App)

	version.Ver().Log()
}

// Reload reload the all configurations
func Reload() error {
	err := archaius.UnmarshalConfig(App)
	if err != nil {
		return err
	}

	*Server = loadServerConfig()
	if GetLog().LogLevel == "DEBUG" {
		body, _ := json.MarshalIndent(archaius.GetConfigs(), "", "  ")
		log.Debug(fmt.Sprintf("finish to reload configurations\n%s", body))
	}

	grcCfg := loadGrcConfig()
	if grcCfg != nil {
		App.Gov.MatchGroup = grcCfg.MatchGroup
		App.Gov.Policies = grcCfg.Policies
	}
	return nil
}

func loadServerConfig() ServerConfig {
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

	return ServerConfig{
		Version: InitVersion,
		// compatible with beego config's runmode
		Environment: GetString("environment", "dev", WithStandby("runmode")),
		Config: ServerConfigDetail{
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

			SslEnabled: GetBool("ssl.enable", true, WithStandby("ssl_mode")),

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

			GlobalVisible: GetString("registry.service.globalVisible", "", WithENV("CSE_SHARED_SERVICES")),
			InstanceTTL:   GetInt64("registry.instance.ttl", 0, WithENV("INSTANCE_TTL")),

			SchemaDisable: GetBool("registry.schema.disable", false, WithENV("SCHEMA_DISABLE")),

			EnableRBAC: GetBool("rbac.enable", false, WithStandby("rbac_enabled")),
		},
	}
}

func setCPUs() {
	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)
	log.Info(fmt.Sprintf("service center is running simultaneously with %d CPU cores", cores))
}

func loadGrcConfig() *Gov {
	grcFile := filepath.Join(util.GetAppRoot(), "conf", "grc.json")
	bytes, err := os.ReadFile(grcFile)
	if err != nil {
		log.Warn(fmt.Sprintf("can not add grc config file source, error: %s", err))
		return nil
	}
	var policies AppConfig
	err = json.Unmarshal(bytes, &policies)
	if err != nil {
		log.Error("unmarshal grc config failed", err)
		return nil
	}
	if policies.Gov == nil {
		log.Warn("no grc config load")
		return nil
	}
	return policies.Gov
}

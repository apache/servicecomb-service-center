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

package broker

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"path/filepath"
)

func Init() {
	//define Broker logger
	name := ""
	if len(config.GetLog().LogFilePath) != 0 {
		name = filepath.Join(filepath.Dir(config.GetLog().LogFilePath), "broker_srvc.log")
	}
	PactLogger = log.NewLogger(log.Config{
		LoggerLevel:    config.GetLog().LogLevel,
		LoggerFile:     name,
		LogFormatText:  config.GetLog().LogFormat == "text",
		LogRotateSize:  int(config.GetLog().LogRotateSize),
		LogBackupCount: int(config.GetLog().LogBackupCount),
	})
	PARTICIPANT = kv.Store().MustInstall(kv.NewAddOn("PARTICIPANT",
		sd.Configure().WithPrefix(GetBrokerParticipantKey(""))))
	VERSION = kv.Store().MustInstall(kv.NewAddOn("VERSION",
		sd.Configure().WithPrefix(GetBrokerVersionKey(""))))
	PACT = kv.Store().MustInstall(kv.NewAddOn("PACT",
		sd.Configure().WithPrefix(GetBrokerPactKey(""))))
	PactVersion = kv.Store().MustInstall(kv.NewAddOn("PACT_VERSION",
		sd.Configure().WithPrefix(GetBrokerPactVersionKey(""))))
	PactTag = kv.Store().MustInstall(kv.NewAddOn("PACT_TAG",
		sd.Configure().WithPrefix(GetBrokerTagKey(""))))
	VERIFICATION = kv.Store().MustInstall(kv.NewAddOn("VERIFICATION",
		sd.Configure().WithPrefix(GetBrokerVerificationKey(""))))
	PactLatest = kv.Store().MustInstall(kv.NewAddOn("PACT_LATEST",
		sd.Configure().WithPrefix(GetBrokerLatestKey(""))))
}

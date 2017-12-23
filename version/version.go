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
package version

import "github.com/astaxie/beego"

var (
	// no need to modify
	// please use:
	// 	go build -ldflags "-X github.com/apache/incubator-servicecomb-service-center/version.VERSION=x.x.x"
	// to set these values.
	VERSION   = "0.0.1"
	BUILD_TAG = "Not provided"
)

type VersionSet struct {
	Version  string `json:"version"`
	BuildTag string `json:"buildTag"`
	RunMode  string `json:"runMode"`
}

var version VersionSet

func init() {
	version.Version = VERSION
	version.BuildTag = BUILD_TAG
	version.RunMode = beego.AppConfig.DefaultString("runmode", "prod")
}

func Ver() *VersionSet {
	return &version
}

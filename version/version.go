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

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"runtime"
)

var (
	// no need to modify
	// please use:
	// 	go build -ldflags "-X github.com/apache/incubator-servicecomb-service-center/version.VERSION=x.x.x"
	// to set these values.
	VERSION   = "0.0.1"
	BUILD_TAG = "Not provided"
)

type VersionSet struct {
	Version   string `json:"version"`
	BuildTag  string `json:"buildTag"`
	GoVersion string `json:"goVersion"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

func (vs *VersionSet) Print() {
	fmt.Printf("ServiceCenter version: %s\n", versionSet.Version)
	fmt.Printf("Build tag: %s\n", versionSet.BuildTag)
	fmt.Printf("Go version: %s\n", versionSet.GoVersion)
	fmt.Printf("OS/Arch: %s/%s\n", versionSet.OS, versionSet.Arch)
}

func (vs *VersionSet) Log() {
	log.Infof("service center version: %s", versionSet.Version)
	log.Infof("Build tag: %s", versionSet.BuildTag)
	log.Infof("Go version: %s", versionSet.GoVersion)
	log.Infof("OS/Arch: %s/%s", versionSet.OS, versionSet.Arch)
}

var versionSet VersionSet

func init() {
	versionSet.Version = VERSION
	versionSet.BuildTag = BUILD_TAG
	versionSet.GoVersion = runtime.Version()
	versionSet.OS = runtime.GOOS
	versionSet.Arch = runtime.GOARCH
}

func Ver() *VersionSet {
	return &versionSet
}

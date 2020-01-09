// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proto

import (
	scerr "github.com/apache/servicecomb-service-center/server/error"
)

const (
	EVT_INIT   EventType = "INIT"
	EVT_CREATE EventType = "CREATE"
	EVT_UPDATE EventType = "UPDATE"
	EVT_DELETE EventType = "DELETE"
	EVT_EXPIRE EventType = "EXPIRE"
	EVT_ERROR  EventType = "ERROR"
	MS_UP      string    = "UP"
	MS_DOWN    string    = "DOWN"

	MSI_UP           string = "UP"
	MSI_DOWN         string = "DOWN"
	MSI_STARTING     string = "STARTING"
	MSI_TESTING      string = "TESTING"
	MSI_OUTOFSERVICE string = "OUTOFSERVICE"

	CHECK_BY_HEARTBEAT string = "push"
	CHECK_BY_PLATFORM  string = "pull"

	EXISTENCE_MS     string = "microservice"
	EXISTENCE_SCHEMA string = "schema"

	PROP_ALLOW_CROSS_APP = "allowCrossApp"

	Response_SUCCESS int32 = 0

	ENV_DEV    string = "development"
	ENV_TEST   string = "testing"
	ENV_ACCEPT string = "acceptance"
	ENV_PROD   string = "production"

	REGISTERBY_SDK      string = "SDK"
	REGISTERBY_SIDECAR  string = "SIDECAR"
	REGISTERBY_PLATFORM string = "PLATFORM"

	APP_ID  = "default"
	VERSION = "0.0.1"
)

func CreateResponse(code int32, message string) *Response {
	resp := &Response{
		Code:    code,
		Message: message,
	}
	return resp
}

func CreateResponseWithSCErr(err *scerr.Error) *Response {
	return &Response{
		Code:    err.Code,
		Message: err.Detail,
	}
}

func DependenciesToKeys(in []*MicroServiceKey, domainProject string) []*MicroServiceKey {
	for _, value := range in {
		if len(value.Tenant) == 0 {
			value.Tenant = domainProject
		}
	}
	return in
}

func MicroServiceToKey(domainProject string, in *MicroService) *MicroServiceKey {
	return &MicroServiceKey{
		Tenant:      domainProject,
		Environment: in.Environment,
		AppId:       in.AppId,
		ServiceName: in.ServiceName,
		Alias:       in.Alias,
		Version:     in.Version,
	}
}

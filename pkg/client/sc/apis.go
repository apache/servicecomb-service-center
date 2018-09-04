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

package sc

import (
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	"github.com/apache/incubator-servicecomb-service-center/server/admin/model"
	"github.com/apache/incubator-servicecomb-service-center/version"
	"io/ioutil"
	"net/http"
)

const (
	apiVersionURL = "/version"
	apiDumpURL    = "/v4/default/admin/dump"
)

func GetScVersion(scClient *rest.URLClient) (*version.VersionSet, error) {
	var headers = make(http.Header)
	if len(Token) > 0 {
		headers.Set("X-Auth-Token", Token)
	}
	resp, err := scClient.HttpDo(http.MethodGet, Addr+apiVersionURL, headers, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, string(body))
	}

	v := &version.VersionSet{}
	err = json.Unmarshal(body, v)
	if err != nil {
		fmt.Println(string(body))
		return nil, err
	}

	return v, nil
}

func GetScCache(scClient *rest.URLClient) (*model.Cache, error) {
	headers := http.Header{
		"X-Domain-Name": []string{"default"},
	}
	if len(Token) > 0 {
		headers.Set("X-Auth-Token", Token)
	}
	resp, err := scClient.HttpDo(http.MethodGet, Addr+apiDumpURL, headers, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, string(body))
	}

	dump := &model.DumpResponse{}
	err = json.Unmarshal(body, dump)
	if err != nil {
		fmt.Println(string(body))
		return nil, err
	}

	return dump.Cache, nil
}

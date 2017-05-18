//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package util

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestNewConnection(t *testing.T) {
	requestPath := "https://99.1.1.64:6443/api/v1/namespaces/default/pods/ruanbinfeng288888-53k5g"
	caFile, err := ioutil.ReadFile("/srv/kubernetes/ca.crt")
	crtFile, err := ioutil.ReadFile("/srv/kubernetes/kubecfg.crt")
	keyFile, err := ioutil.ReadFile("/srv/kubernetes/kubecfg.key")
	client, err := NewConnection(requestPath, caFile, crtFile, keyFile)
	if err != nil {
		t.Error(err)
		t.Failed()
	}
	req, err := http.NewRequest("GET", requestPath, nil)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		t.Error("send a a request message to k8s err:", err)
		t.Failed()
	} else {
		t.Log("got OK from response")
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("return code is not OK")
		t.Failed()
	}
}

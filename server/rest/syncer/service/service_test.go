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
package service_test

// initialize
import _ "github.com/apache/servicecomb-service-center/test"

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/rest/syncer"
	"github.com/apache/servicecomb-service-center/server/rest/syncer/service"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type mockSyncerHandler struct {
	Func func(w http.ResponseWriter, r *http.Request)
}

func (m *mockSyncerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Func(w, r)
}

func TestSyncerService_WatchInstance(t *testing.T) {
	//create fake config file for test
	defer archaius.Clean()
	b := []byte(`
syncer:
enabled: true
`)
	dir := filepath.Join(util.GetAppRoot(), "conf")
	defer os.Remove(dir)
	os.Mkdir(dir, 0750)
	file := filepath.Join(dir, "app.yaml")
	defer os.Remove(file)
	f1, err := os.Create(file)
	assert.NoError(t, err)
	_, err = io.WriteString(f1, string(b))
	assert.NoError(t, err)
	config.Init()
	assert.NoError(t, err)

	svr := httptest.NewServer(&mockSyncerHandler{func(w http.ResponseWriter, r *http.Request) {
		ctrl := &syncer.SyncerController{}
		ctrl.WatchInstance(w, r.WithContext(getContext()))
	}})
	defer svr.Close()

	// error situation test
	w := httptest.NewRecorder()
	r := &http.Request{}
	service.ServiceAPI.WatchInstance(w, r.WithContext(getContext()))
}

func getContext() context.Context {
	return util.SetContext(
		util.SetDomainProject(context.Background(), "default", "default"),
		util.CtxNocache, "1")
}

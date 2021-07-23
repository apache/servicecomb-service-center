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

package config_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	defer archaius.Clean()
	b := []byte(`
gov:
  kieSever1:
   type: kie
   endpoint: 127.0.0.1:30110

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
	assert.Equal(t, "kie", config.GetGov().DistMap["kieSever1"].Type)
	assert.Equal(t, "127.0.0.1:30110", config.GetGov().DistMap["kieSever1"].Endpoint)
}

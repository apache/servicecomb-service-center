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
	"github.com/apache/servicecomb-service-center/server/core/config"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	b := []byte(`
gov:
  plugins:
    - type: mock
    - type: kie
    - type: istio

`)
	defer os.Remove("test.yaml")
	f1, err := os.Create("test.yaml")
	assert.NoError(t, err)
	_, err = io.WriteString(f1, string(b))
	assert.NoError(t, err)
	config.Configurations.ConfigFile = "test.yaml"
	config.Init()
	assert.NoError(t, err)
	assert.Equal(t, "mock", config.GetGov().DistOptions[0].Type)
}

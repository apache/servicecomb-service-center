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
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestGetString(t *testing.T) {
	os.Setenv("TEST_GET_STRING", "test")
	defer archaius.Clean()
	archaius.Init(archaius.WithMemorySource(), archaius.WithENVSource(),
		archaius.WithOptionalFiles([]string{filepath.Join(util.GetAppRoot(), "conf", "app.yaml")}))
	assert.Equal(t, "test", config.GetString("test.getString", "none"))
	assert.Equal(t, "test", config.GetString("other.getString", "none", config.WithENV("TEST_GET_STRING")))
}

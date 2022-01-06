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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	changeConfigPath()
	assert.NoError(t, config.Init())
	assert.NotNil(t, config.GetConfig().Sync)
}

func changeConfigPath() {
	workDir, _ := os.Getwd()
	replacePath := filepath.Join("syncer", "config")
	workDir = strings.ReplaceAll(workDir, replacePath, "etc")
	os.Setenv("APP_ROOT", workDir)
}

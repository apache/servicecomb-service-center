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

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
)

func TestGetString(t *testing.T) {
	os.Setenv("TEST_GET_STRING", "test")
	defer archaius.Clean()
	archaius.Init(archaius.WithMemorySource(), archaius.WithENVSource(),
		archaius.WithOptionalFiles([]string{filepath.Join(util.GetAppRoot(), "conf", "app.yaml")}))
	assert.Equal(t, "test", config.GetString("test.getString", "none"))
	assert.Equal(t, "test", config.GetString("other.getString", "none", config.WithENV("TEST_GET_STRING")))
}

func TestGetStringMap(t *testing.T) {
	changeConfigPath()
	os.Setenv("TEST_GET_STRING", "test")
	defer archaius.Clean()
	archaius.Init(archaius.WithMemorySource(), archaius.WithENVSource(),
		archaius.WithOptionalFiles([]string{filepath.Join(util.GetAppRoot(), "conf", "app.yaml")}))

	t.Run("has no data from config should be passed", func(t *testing.T) {
		_ = archaius.Set("a.b.c", "test_value_c")
		_ = archaius.Set("a.b.d", "test_value_d")
		maps := config.GetStringMap("a.c")
		assert.Equal(t, 0, len(maps))
	})
	archaius.Clean()
	t.Run("has same key , should be passed", func(t *testing.T) {
		_ = archaius.Set("a.b.c", "test_value_c")
		_ = archaius.Set("a.bb.d", "test_value_d")
		maps := config.GetStringMap("a.b")
		assert.Equal(t, 1, len(maps))
		c, ok := maps["c"]
		assert.Equal(t, true, ok)
		assert.Equal(t, "test_value_c", c)
	})
	archaius.Clean()
	t.Run("has first value, should be passed", func(t *testing.T) {
		_ = archaius.Set("a.b.c", "test_value_c")
		_ = archaius.Set("a.b.d", "test_value_d")
		maps := config.GetStringMap("a")
		assert.Equal(t, 0, len(maps))
	})
	archaius.Clean()
	t.Run("value type is different, should be passed", func(t *testing.T) {
		_ = archaius.Set("aa.b.c", 1)
		_ = archaius.Set("aa.b.d", "test_value_d")
		maps := config.GetStringMap("aa.b")
		assert.Equal(t, 2, len(maps))
		c, ok1 := maps["c"]
		d, ok2 := maps["d"]
		assert.Equal(t, true, ok1 && ok2)
		assert.Equal(t, "1", c)
		assert.Equal(t, "test_value_d", d)
	})

}

func changeConfigPath() {
	workDir, _ := os.Getwd()
	replacePath := filepath.Join("server", "config")
	workDir = strings.ReplaceAll(workDir, replacePath, "etc")
	os.Setenv("APP_ROOT", workDir)
}

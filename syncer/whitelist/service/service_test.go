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

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/whitelist/service"
)

func TestEnableSync(t *testing.T) {

	t.Run("Sync switch off", func(t *testing.T) {
		syncSwitchOffConfig()
		t.Run("The sync switch is turned off and no matter what the service name is, it returns a failure", func(t *testing.T) {
			assert.Equal(t, false, service.IsExistInWhiteList("AAA"))
			assert.Equal(t, false, service.IsExistInWhiteList("BBB"))
			assert.Equal(t, false, service.IsExistInWhiteList("CCC"))
		})
	})

	t.Run("Sync switch on", func(t *testing.T) {
		syncSwitchOnConfig()

		t.Run("The service name is in the whitelist will sync successfully", func(t *testing.T) {
			assert.Equal(t, true, service.IsExistInWhiteList("AAA"))
			assert.Equal(t, true, service.IsExistInWhiteList("BBB"))
		})

		t.Run("The service name is not in the whitelist will sync failed", func(t *testing.T) {
			assert.Equal(t, false, service.IsExistInWhiteList("CCC"))
		})
	})

	t.Run("Sync switch on without rules", func(t *testing.T) {
		syncSwitchOnWithoutRulesConfig()
		// compatible with older versions
		assert.Equal(t, true, service.IsExistInWhiteList("AAA"))
	})
}

func syncSwitchOnConfig() {
	cfg := config.Config{
		Sync: &config.Sync{
			EnableOnStart: true,
			WhiteList: &config.WhiteList{
				Service: &config.Service{
					Rules: []string{"A*", "B*"},
				},
			},
		},
	}
	config.SetConfig(cfg)
}

func syncSwitchOffConfig() {
	cfg := config.Config{
		Sync: &config.Sync{
			EnableOnStart: false,
			WhiteList: &config.WhiteList{
				Service: &config.Service{
					Rules: []string{"A*", "B*"},
				},
			},
		},
	}
	config.SetConfig(cfg)
}

func syncSwitchOnWithoutRulesConfig() {
	cfg := config.Config{
		Sync: &config.Sync{
			EnableOnStart: true,
			WhiteList: &config.WhiteList{
				Service: &config.Service{
					Rules: []string{},
				},
			},
		},
	}
	config.SetConfig(cfg)
}

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

package account_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/whitelist/account"
)

func TestEnableSync(t *testing.T) {

	t.Run("Sync switch off", func(t *testing.T) {
		syncSwitchOffConfig()
		t.Run("The sync switch is turned off and no matter what the account name is, it returns a failure", func(t *testing.T) {
			assert.Equal(t, false, account.IsExistInWhiteList("AAA"))
			assert.Equal(t, false, account.IsExistInWhiteList("BBB"))
			assert.Equal(t, false, account.IsExistInWhiteList("CCC"))
		})
	})

	t.Run("Sync switch on", func(t *testing.T) {
		syncSwitchOnConfig()

		t.Run("The account name is in the whitelist will sync successfully", func(t *testing.T) {
			assert.Equal(t, true, account.IsExistInWhiteList("AAA"))
			assert.Equal(t, true, account.IsExistInWhiteList("BBB"))
		})

		t.Run("The account name is not in the whitelist will sync failed", func(t *testing.T) {
			assert.Equal(t, false, account.IsExistInWhiteList("CCC"))
		})
	})

	t.Run("Sync switch on without rules", func(t *testing.T) {
		syncSwitchOnWithoutRulesConfig()
		// there is no white list for the synchronization of compatible old versions
		assert.Equal(t, true, account.IsExistInWhiteList("AAA"))
	})

}

func syncSwitchOnConfig() {
	cfg := config.Config{
		Sync: &config.Sync{
			EnableOnStart: true,
			WhiteList: &config.WhiteList{
				Account: &config.Account{
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
				Account: &config.Account{
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
				Account: &config.Account{
					Rules: []string{},
				},
			},
		},
	}
	config.SetConfig(cfg)
}

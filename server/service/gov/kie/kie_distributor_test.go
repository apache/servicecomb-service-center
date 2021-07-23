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

package kie_test

import (
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/server/config"
	govsvc "github.com/apache/servicecomb-service-center/server/service/gov"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	App     = "test1"
	Env0    = "env0"
	Env1    = "env1"
	Project = "test_kie"
)

func init() {
	config.App.Gov = &config.Gov{
		DistMap: map[string]config.DistributorOptions{
			"kie": {
				Type:     "kie",
				Endpoint: "http://127.0.0.1:27017",
			},
		},
	}
	err := govsvc.Init()
	if err != nil {
		panic(err)
	}
}

func TestDeleteMatchGroup(t *testing.T) {
	var id1 string
	var id2 string
	t.Run("create two match-group with the same name and different environments", func(t *testing.T) {
		res, err := govsvc.Create("match-group", Project, &gov.Policy{
			GovernancePolicy: &gov.GovernancePolicy{
				Name: "test",
				Selector: &gov.Selector{
					App:         App,
					Environment: Env0,
				},
			},
			Spec: map[string]interface{}{"retryNext": 3, "match": "test"},
		})
		id1 = string(res)
		assert.NoError(t, err)
		assert.NotEmpty(t, id1)

		res, err = govsvc.Create("match-group", Project, &gov.Policy{
			GovernancePolicy: &gov.GovernancePolicy{
				Name: "test",
				Selector: &gov.Selector{
					App:         App,
					Environment: Env1,
				},
			},
			Spec: map[string]interface{}{"retryNext": 3, "match": "traffic2adminAPI"},
		})
		id2 = string(res)
		assert.NoError(t, err)
		assert.NotEmpty(t, id2)
	})
	t.Run("delete one of match-group, should delete 1 items", func(t *testing.T) {
		err := govsvc.Delete("match-group", id1, Project)
		assert.NoError(t, err)
		res1, _ := govsvc.Get("match-group", id1, Project)
		assert.Nil(t, res1)

		res2, _ := govsvc.Get("match-group", id2, Project)
		assert.NotNil(t, res2)

		err = govsvc.Delete("match-group", id2, Project)
		assert.NoError(t, err)
		res2, _ = govsvc.Get("match-group", id2, Project)
		fmt.Println(res2)
	})
}

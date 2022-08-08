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

package grc_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/server/config"
	grcsvc "github.com/apache/servicecomb-service-center/server/service/grc"
	_ "github.com/apache/servicecomb-service-center/server/service/grc/mock"
)

const Project = "default"
const MockKind = "default"
const MatchGroup = "match-group"
const MockEnv = ""
const MockApp = ""

var id = ""

func init() {
	config.App = &config.AppConfig{
		Gov: &config.Gov{
			DistMap: map[string]config.DistributorOptions{
				"mock": {
					Type: "mock",
				},
			},
		},
	}
	err := grcsvc.Init()
	if err != nil {
		panic(err)
	}
}

func TestCreate(t *testing.T) {
	res, err := grcsvc.Create(context.TODO(), MockKind, Project, &gov.Policy{
		GovernancePolicy: &gov.GovernancePolicy{
			Name: "Traffic2adminAPI",
			Selector: gov.Selector{
				"app":         MockApp,
				"environment": MockEnv,
			},
		},
		Spec: map[string]interface{}{"retryNext": 3, "match": "traffic2adminAPI"},
	})
	id = string(res)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestUpdate(t *testing.T) {
	err := grcsvc.Update(context.TODO(), MockKind, id, Project, &gov.Policy{
		GovernancePolicy: &gov.GovernancePolicy{
			Name: "Traffic2adminAPI",
			Selector: gov.Selector{
				"app":         MockApp,
				"environment": MockEnv,
			},
		},
		Spec: map[string]interface{}{"retryNext": 3, "match": "traffic2adminAPI"},
	})
	assert.NoError(t, err)
}

func TestDisplay(t *testing.T) {
	res, err := grcsvc.Create(context.TODO(), MatchGroup, Project, &gov.Policy{
		GovernancePolicy: &gov.GovernancePolicy{
			Name: "Traffic2adminAPI",
			Selector: gov.Selector{
				"app":         MockApp,
				"environment": MockEnv,
			},
		},
	})
	id = string(res)
	assert.NoError(t, err)
	policies := &[]*gov.DisplayData{}
	res, err = grcsvc.Display(context.TODO(), Project, MockApp, MockEnv)
	assert.NoError(t, err)
	err = json.Unmarshal(res, policies)
	assert.NoError(t, err)
	assert.NotEmpty(t, policies)
}

func TestList(t *testing.T) {
	policies := &[]*gov.Policy{}
	res, err := grcsvc.List(context.TODO(), MockKind, Project, MockApp, MockEnv)
	assert.NoError(t, err)
	err = json.Unmarshal(res, policies)
	assert.NoError(t, err)
	assert.NotEmpty(t, policies)
}

func TestGet(t *testing.T) {
	policy := &gov.Policy{}
	res, err := grcsvc.Get(context.TODO(), MockKind, id, Project)
	assert.NoError(t, err)
	err = json.Unmarshal(res, policy)
	assert.NoError(t, err)
	assert.NotNil(t, policy)
}

func TestDelete(t *testing.T) {
	err := grcsvc.Delete(context.TODO(), MockKind, id, Project)
	assert.NoError(t, err)
	res, _ := grcsvc.Get(context.TODO(), MockKind, id, Project)
	assert.Nil(t, res)
}

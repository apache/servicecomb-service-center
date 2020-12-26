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

package gov_test

import (
	"encoding/json"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/server/config"
	svc "github.com/apache/servicecomb-service-center/server/service/gov"
	_ "github.com/apache/servicecomb-service-center/server/service/gov/mock"
	"github.com/stretchr/testify/assert"
)

const Project = "default"

const MockKind = "default"

func init() {
	config.Configurations = &config.Config{
		Gov: &config.Gov{
			DistOptions: []config.DistributorOptions{
				{
					Name: "mockServer",
					Type: "mock",
				},
			},
		},
	}
	err := svc.Init()
	panic(err)
}

func TestCreate(t *testing.T) {
	b, _ := json.MarshalIndent(&gov.Policy{
		GovernancePolicy: &gov.GovernancePolicy{
			Name: "Traffic2adminAPI",
		},
		Spec: &gov.LBSpec{RetryNext: 3, MarkerName: "traffic2adminAPI"},
	}, "", "  ")
	_, err := svc.Create(MockKind, Project, b)
	assert.NoError(t, err)
}

func TestUpdate(t *testing.T) {
	b, _ := json.MarshalIndent(&gov.Policy{
		GovernancePolicy: &gov.GovernancePolicy{
			Name: "Traffic2adminAPI",
		},
		Spec: &gov.LBSpec{RetryNext: 3, MarkerName: "traffic2adminAPI"},
	}, "", "  ")
	err := svc.Update("xxxxxx", MockKind, Project, b)
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
}

func TestDisplay(t *testing.T) {
}

func TestList(t *testing.T) {
	//policies := &[]*gov.Policy{}
	//res, err := svc.List(MockKind, Project, "", "")
	//panic(err)
	//err = json.Unmarshal(res, policies)
	//panic(err)
}

func TestGet(t *testing.T) {
}

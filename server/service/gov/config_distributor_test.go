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
	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/server/config"
	svc "github.com/apache/servicecomb-service-center/server/service/gov"
	_ "github.com/apache/servicecomb-service-center/server/service/gov/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreate(t *testing.T) {
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
	assert.NoError(t, err)
	b, _ := json.MarshalIndent(&gov.LoadBalancer{
		GovernancePolicy: &gov.GovernancePolicy{
			Name: "Traffic2adminAPI",
		},
		Spec: &gov.LBSpec{RetryNext: 3, MarkerName: "traffic2adminAPI"},
	}, "", "  ")
	err = svc.Create("lb", "default", b)
	assert.NoError(t, err)
}

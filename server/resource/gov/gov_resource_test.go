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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/apache/servicecomb-service-center/server/service/grc/mock"
	"k8s.io/kube-openapi/pkg/validation/spec"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/config"
	v1 "github.com/apache/servicecomb-service-center/server/resource/gov"
	grcsvc "github.com/apache/servicecomb-service-center/server/service/grc"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
)

func init() {
	config.App.Gov = &config.Gov{
		Policies: config.Policies{
			"mock": {
				ValidationSpec: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type:     []string{"object"},
						Required: []string{"rule"},
					},
				},
			},
		},
	}
	err := grcsvc.Init()
	if err != nil {
		log.Fatal("", err)
	}
}

func TestGovernance_Create(t *testing.T) {
	err := archaius.Init(archaius.WithMemorySource(), archaius.WithENVSource())
	assert.NoError(t, err)

	rest.RegisterServant(&v1.Governance{})

	t.Run("create load balancing, should success", func(t *testing.T) {
		b, _ := json.Marshal(&gov.Policy{
			GovernancePolicy: &gov.GovernancePolicy{Name: "test"},
			Spec: map[string]interface{}{
				"rule": "rr",
			}})
		r, _ := http.NewRequest(http.MethodPost, "/v1/default/gov/mock", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("create load balancing without rule, should failed", func(t *testing.T) {
		b, _ := json.Marshal(&gov.Policy{
			GovernancePolicy: &gov.GovernancePolicy{Name: "test"},
			Spec:             map[string]interface{}{}})
		r, _ := http.NewRequest(http.MethodPost, "/v1/default/gov/mock", bytes.NewBuffer(b))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

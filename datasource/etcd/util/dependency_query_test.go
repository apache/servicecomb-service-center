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

package util

import (
	"context"
	"errors"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestDependency(t *testing.T) {
	d := &Dependency{
		DeleteDependencyRuleList: []*discovery.MicroServiceKey{
			{ServiceName: "b", Version: "1.0.0"},
		},
		CreateDependencyRuleList: []*discovery.MicroServiceKey{
			{ServiceName: "a", Version: "1.0.0"},
		},
	}
	err := d.Commit(context.Background())
	if err != nil {
		t.Fatalf(`Dependency_UpdateProvidersRuleOfConsumer failed`)
	}

	dr := NewDependencyRelation(context.Background(), "", &discovery.MicroService{}, &discovery.MicroService{})
	_, err = dr.GetProviderIdsByRules([]*discovery.MicroServiceKey{
		{ServiceName: "*"},
	})
	if err != nil {
		t.Fatalf(`DependencyRelation_getDependencyProviderIds * failed`)
	}
	_, err = dr.GetProviderIdsByRules([]*discovery.MicroServiceKey{
		{ServiceName: "a", Version: "1.0.0"},
		{ServiceName: "b", Version: "latest"},
	})
	if err != nil {
		t.Fatalf(`DependencyRelation_getDependencyProviderIds failed`)
	}

	_, err = dr.GetDependencyConsumers()
	if err != nil {
		t.Fatalf(`DependencyRelation_GetDependencyConsumers failed`)
	}

	t.Run("GetServiceByMicroServiceKey when service not exist", func(t *testing.T) {
		_, err = dr.GetServiceByMicroServiceKey(&discovery.MicroServiceKey{})
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			t.Fatalf(`DependencyRelation_getServiceByMicroServiceKey failed`)
		}
		assert.Equal(t, datasource.ErrNoData, err, "service not exist")
	})

	_, err = dr.GetConsumerOfSameServiceNameAndAppID(&discovery.MicroServiceKey{})
	if err != nil {
		t.Fatalf(`DependencyRelation_getConsumerOfSameServiceNameAndAppId failed`)
	}

	dr = NewConsumerDependencyRelation(context.Background(), "", &discovery.MicroService{})
	_, err = dr.GetDependencyConsumersOfProvider()
	if err == nil {
		t.Fatalf(`DependencyRelation_getDependencyConsumersOfProvider failed`)
	}
	_, err = dr.GetDependencyProviders()
	if err != nil {
		t.Fatalf(`DependencyRelation_GetDependencyProviders failed`)
	}
}

func TestDependencyRelationFilterOpt(t *testing.T) {
	op := ToDependencyRelationFilterOpt(
		WithSameDomainProject(),
		WithoutSelfDependency(),
	)
	if !op.NonSelf || !op.SameDomainProject {
		t.Fatalf(`ToDependencyRelationFilterOpt failed`)
	}
}

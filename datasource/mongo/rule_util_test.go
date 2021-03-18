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

package mongo_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
)

func TestRuleFilter_Filter(t *testing.T) {
	var err error
	t.Run("when there is no such a customer in db", func(t *testing.T) {
		_, err = mongo.Filter(context.Background(), []*model.Rule{}, "")
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			t.Fatalf("RuleFilter Filter failed")
		}
		assert.Equal(t, datasource.ErrNoData, err, "no data found")
	})
	t.Run("FilterAll when customer not exist", func(t *testing.T) {
		_, _, err = mongo.FilterAll(context.Background(), []string{""}, []*model.Rule{})
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			t.Fatalf("RuleFilter FilterAll failed")
		}
		assert.Equal(t, nil, err, "no customer found err is nil")
	})
	t.Run("FilterAll when ProviderRules not nil and service not exist", func(t *testing.T) {
		_, _, err = mongo.FilterAll(context.Background(), []string{""}, []*model.Rule{{}})
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			t.Fatalf("RuleFilter FilterAll failed")
		}
		assert.Equal(t, datasource.ErrNoData, err, "no customer found when FilterAll")
	})
}

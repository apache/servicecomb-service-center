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

package quota_test

import (
	"context"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/util"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/stretchr/testify/assert"
)

func TestApplyService(t *testing.T) {
	//var id string
	ctx := context.TODO()
	ctx = util.SetDomainProject(ctx, "quota", "quota")
	t.Run("create service, should success", func(t *testing.T) {
		err := quotasvc.ApplyService(ctx, 1)
		assert.Nil(t, err)
	})
}

func TestRemandService(t *testing.T) {
	quotasvc.RemandService(context.Background())
}

// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package buildin_test

import (
	"context"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/util"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestGetResourceLimit(t *testing.T) {
	//var id string
	ctx := context.TODO()
	ctx = util.SetDomainProject(ctx, "quota", "quota")
	t.Run("create service,should success", func(t *testing.T) {
		err := quotasvc.ApplyService(ctx, 1)
		assert.Nil(t, err)
	})
	t.Run("create 1 instance,should success", func(t *testing.T) {
		resp, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "quota",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		err = quotasvc.ApplyInstance(ctx, 1)
		assert.Nil(t, err)

		err = quotasvc.ApplyInstance(ctx, 150001)
		assert.NotNil(t, err)
	})

	t.Run("create 150001 instance,should failed", func(t *testing.T) {
		resp, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "quota2",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, pb.ResponseSuccess, resp.Response.GetCode())

		err = quotasvc.ApplyInstance(ctx, 150001)
		assert.NotNil(t, err)
	})
}

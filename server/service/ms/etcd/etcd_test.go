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

package etcd_test

import (
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/service/ms"
	"github.com/apache/servicecomb-service-center/server/service/ms/etcd"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	TooLongAppId       = strings.Repeat("x", 162)
	TooLongSchemaId    = strings.Repeat("x", 162)
	TooLongServiceName = strings.Repeat("x", 130)
	TooLongAlias       = strings.Repeat("x", 130)
	TooLongFramework   = strings.Repeat("x", 66)
)

func TestInit(t *testing.T) {
	_ = archaius.Init(archaius.WithMemorySource())
	_ = archaius.Set("servicecomb.ms.name", "etcd")

	t.Run("Register service after init & install, should pass", func(t *testing.T) {
		ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
			return etcd.NewDataSource(), nil
		})

		err := ms.Init(ms.Options{
			Endpoint:       "",
			PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
		})
		assert.NoError(t, err)

		size := quota.DefaultSchemaQuota + 1
		paths := make([]*pb.ServicePath, 0, size)
		properties := make(map[string]string, size)
		for i := 0; i < size; i++ {
			s := strconv.Itoa(i) + strings.Repeat("x", 253)
			paths = append(paths, &pb.ServicePath{Path: s, Property: map[string]string{s: s}})
			properties[s] = s
		}
		request := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				AppId:       TooLongAppId[:len(TooLongAppId)-1],
				ServiceName: TooLongServiceName[:len(TooLongServiceName)-1],
				Version:     "32767.32767.32767.32767",
				Alias:       TooLongAlias[:len(TooLongAlias)-1],
				Level:       "BACK",
				Status:      "UP",
				Schemas:     []string{TooLongSchemaId[:len(TooLongSchemaId)-1]},
				Paths:       paths,
				Properties:  properties,
				Framework: &pb.FrameWorkProperty{
					Name:    TooLongFramework[:len(TooLongFramework)-1],
					Version: TooLongFramework[:len(TooLongFramework)-1],
				},
				RegisterBy: "SDK",
				Timestamp:  strconv.FormatInt(time.Now().Unix(), 10),
			},
		}
		request.Service.ModTimestamp = request.Service.Timestamp
		resp, err := ms.MicroService().RegisterService(getContext(), request)

		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.True(t, resp.Succeeded)
	})

	t.Run("register service with framework name nil", func(t *testing.T) {
		ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
			return etcd.NewDataSource(), nil
		})

		err := ms.Init(ms.Options{
			Endpoint:       "",
			PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
		})
		assert.NoError(t, err)

		request := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "framework-test-ms",
				AppId:       "default",
				Version:     "1.0.1",
				Level:       "BACK",
				Framework: &pb.FrameWorkProperty{
					Version: "1.0.0",
				},
				Properties: make(map[string]string),
				Status:     "UP",
			},
		}

		resp, err := ms.MicroService().RegisterService(getContext(), request)

		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.True(t, resp.Succeeded)
	})

	t.Run("register service with framework version nil", func(t *testing.T) {
		ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
			return etcd.NewDataSource(), nil
		})
		err := ms.Init(ms.Options{
			Endpoint:       "",
			PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
		})
		assert.NoError(t, err)
		request := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "framework-test-ms",
				AppId:       "default",
				Version:     "1.0.2",
				Level:       "BACK",
				Framework: &pb.FrameWorkProperty{
					Name: "framework",
				},
				Properties: make(map[string]string),
				Status:     "UP",
			},
		}

		resp, err := ms.MicroService().RegisterService(getContext(), request)

		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.True(t, resp.Succeeded)
	})

	t.Run("register service with status is nil", func(t *testing.T) {
		ms.Install("etcd", func(opts ms.Options) (ms.DataSource, error) {
			return etcd.NewDataSource(), nil
		})
		err := ms.Init(ms.Options{
			Endpoint:       "",
			PluginImplName: ms.ImplName(archaius.GetString("servicecomb.ms.name", "etcd")),
		})
		assert.NoError(t, err)

		request := &pb.CreateServiceRequest{
			Service: &pb.MicroService{
				ServiceName: "status-test-ms",
				AppId:       "default",
				Version:     "1.0.3",
				Level:       "BACK",
				Properties:  make(map[string]string),
			},
		}

		resp, err := ms.MicroService().RegisterService(getContext(), request)

		assert.NotNil(t, resp)
		assert.NoError(t, err)
		assert.True(t, resp.Succeeded)
	})
}

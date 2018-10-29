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
package backend

import (
	"errors"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/gopool"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/registry"
	"golang.org/x/net/context"
	"sync"
	"time"
)

var (
	engineInstance *registryEngine
	singletonLock  sync.Mutex
)

const (
	// the same as v3rpc.MaxOpsPerTxn = 128
	MAX_TXN_NUMBER_ONE_TIME = 128
)

func NewEngine() (*registryEngine, error) {
	instance := plugin.Plugins().Registry()
	if instance == nil {
		return nil, errors.New("register center client plugin does not exist")
	}
	select {
	case err := <-instance.Err():
		plugin.Plugins().Reload(plugin.REGISTRY)
		return nil, err
	case <-instance.Ready():
	}
	return &registryEngine{
		Registry:  instance,
		goroutine: gopool.New(context.Background()),
	}, nil
}

func Registry() registry.Registry {
	return RegistryEngine()
}

func RegistryEngine() *registryEngine {
	if engineInstance == nil {
		singletonLock.Lock()
		for i := 0; engineInstance == nil; i++ {
			inst, err := NewEngine()
			if err != nil {
				log.Errorf(err, "get register center client failed")
			}
			engineInstance = inst

			if engineInstance != nil {
				singletonLock.Unlock()
				return engineInstance
			}

			t := util.GetBackoff().Delay(i)
			log.Errorf(nil, "initialize service center failed, retry after %s", t)
			<-time.After(t)
		}
		singletonLock.Unlock()
	}
	return engineInstance
}

func BatchCommit(ctx context.Context, opts []registry.PluginOp) error {
	_, err := BatchCommitWithCmp(ctx, opts, nil, nil)
	return err
}

func BatchCommitWithCmp(ctx context.Context, opts []registry.PluginOp,
	cmp []registry.CompareOp, fail []registry.PluginOp) (resp *registry.PluginResponse, err error) {
	lenOpts := len(opts)
	tmpLen := lenOpts
	tmpOpts := []registry.PluginOp{}
	for i := 0; tmpLen > 0; i++ {
		tmpLen = lenOpts - (i+1)*MAX_TXN_NUMBER_ONE_TIME
		if tmpLen > 0 {
			tmpOpts = opts[i*MAX_TXN_NUMBER_ONE_TIME : (i+1)*MAX_TXN_NUMBER_ONE_TIME]
		} else {
			tmpOpts = opts[i*MAX_TXN_NUMBER_ONE_TIME : lenOpts]
		}
		resp, err = Registry().TxnWithCmp(ctx, tmpOpts, cmp, fail)
		if err != nil || !resp.Succeeded {
			return
		}
	}
	return
}

type registryEngine struct {
	registry.Registry
	goroutine *gopool.Pool
}

func (s *registryEngine) Start() error {
	err := s.selfRegister(context.Background())
	if err != nil {
		return err
	}

	s.heartBeatService()
	ReportScInstance()
	return nil
}

func (s *registryEngine) Stop() {
	s.goroutine.Close(true)

	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	s.unregisterInstance(ctx)
}

func (s *registryEngine) selfRegister(ctx context.Context) error {
	err := s.registryService(context.Background())
	if err != nil {
		return err
	}
	// 实例信息
	return s.registryInstance(context.Background())
}

func (s *registryEngine) registryService(pCtx context.Context) error {
	ctx := core.AddDefaultContextValue(pCtx)
	respE, err := core.ServiceAPI.Exist(ctx, core.GetExistenceRequest())
	if err != nil {
		log.Error("query service center existence failed", err)
		return err
	}
	if respE.Response.Code == pb.Response_SUCCESS {
		log.Warnf("service center service[%s] already registered", respE.ServiceId)
		respG, err := core.ServiceAPI.GetOne(ctx, core.GetServiceRequest(respE.ServiceId))
		if respG.Response.Code != pb.Response_SUCCESS {
			log.Errorf(err, "query service center service[%s] info failed", respE.ServiceId)
			return fmt.Errorf("service center service file lost.")
		}
		core.Service = respG.Service
		return nil
	}

	respS, err := core.ServiceAPI.Create(ctx, core.CreateServiceRequest())
	if err != nil {
		log.Error("register service center failed", err)
		return err
	}
	core.Service.ServiceId = respS.ServiceId
	log.Infof("register service center service[%s]", respS.ServiceId)
	return nil
}

func (s *registryEngine) registryInstance(pCtx context.Context) error {
	core.Instance.InstanceId = ""
	core.Instance.ServiceId = core.Service.ServiceId

	ctx := core.AddDefaultContextValue(pCtx)

	respI, err := core.InstanceAPI.Register(ctx, core.RegisterInstanceRequest())
	if respI.Response.Code != pb.Response_SUCCESS {
		err = fmt.Errorf("register service center[%s] instance failed, %s",
			core.Instance.ServiceId, respI.Response.Message)
		log.Error(err.Error(), nil)
		return err
	}
	core.Instance.InstanceId = respI.InstanceId
	log.Infof("register service center instance[%s/%s], endpoints is %s",
		core.Service.ServiceId, respI.InstanceId, core.Instance.Endpoints)
	return nil
}

func (s *registryEngine) unregisterInstance(pCtx context.Context) error {
	if len(core.Instance.InstanceId) == 0 {
		return nil
	}
	ctx := core.AddDefaultContextValue(pCtx)
	respI, err := core.InstanceAPI.Unregister(ctx, core.UnregisterInstanceRequest())
	if respI.Response.Code != pb.Response_SUCCESS {
		err = fmt.Errorf("unregister service center instance[%s/%s] failed, %s",
			core.Instance.ServiceId, core.Instance.InstanceId, respI.Response.Message)
		log.Error(err.Error(), nil)
		return err
	}
	log.Warnf("unregister service center instance[%s/%s]",
		core.Service.ServiceId, core.Instance.InstanceId)
	return nil
}

func (s *registryEngine) sendHeartBeat(pCtx context.Context) {
	ctx := core.AddDefaultContextValue(pCtx)
	respI, err := core.InstanceAPI.Heartbeat(ctx, core.HeartbeatRequest())
	if respI.Response.Code == pb.Response_SUCCESS {
		log.Debugf("update service center instance[%s/%s] heartbeat",
			core.Instance.ServiceId, core.Instance.InstanceId)
		return
	}
	log.Errorf(err, "update service center instance[%s/%s] heartbeat failed",
		core.Instance.ServiceId, core.Instance.InstanceId)

	//服务不存在，创建服务
	err = s.selfRegister(pCtx)
	if err != nil {
		log.Errorf(err, "retry to register[%s/%s/%s/%s] failed",
			core.Service.Environment, core.Service.AppId, core.Service.ServiceName, core.Service.Version)
	}
}

func (s *registryEngine) heartBeatService() {
	s.goroutine.Do(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(core.Instance.HealthCheck.Interval) * time.Second):
				s.sendHeartBeat(ctx)
			}
		}
	})
}

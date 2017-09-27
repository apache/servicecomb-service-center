//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//Test Bot functionality
package server

import (
	"fmt"
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	rs "github.com/ServiceComb/service-center/server/rest"
	"github.com/ServiceComb/service-center/server/rpc"
	"github.com/ServiceComb/service-center/server/service"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/util/grace"
	"github.com/ServiceComb/service-center/util/rest"
	"golang.org/x/net/context"
	"time"
)

var apiServer *APIServer

func init() {
	InitAPI()
}

func InitAPI() {
	core.ServiceAPI, core.InstanceAPI, core.GovernServiceAPI = service.AssembleResources()
	apiServer = &APIServer{
		isClose: true,
		err:     make(chan error, 1),
	}
}

type APIType int64

func (t APIType) String() string {
	switch t {
	case RPC:
		return "grpc" // support grpc
	case REST:
		return "rest"
	default:
		return fmt.Sprintf("SCHEME%d", t)
	}
}

type APIServer struct {
	HostName  string
	Endpoints map[APIType]string
	restSrv   *rest.Server
	rpcSrv    *rpc.Server
	isClose   bool
	forked    bool
	err       chan error
}

const (
	RPC  APIType = 0
	REST APIType = 1
)

func (s *APIServer) Err() <-chan error {
	return s.err
}

func (s *APIServer) registerServiceCenter() error {
	err := s.registryService(context.Background())
	if err != nil {
		return err
	}
	// 实例信息
	return s.registryInstance(context.Background())
}

func (s *APIServer) registryService(pCtx context.Context) error {
	ctx := core.AddDefaultContextValue(pCtx)
	respE, err := core.ServiceAPI.Exist(ctx, core.GetExistenceRequest())
	if err != nil {
		util.Logger().Error("query service center existence failed", err)
		return err
	}
	if respE.Response.Code == pb.Response_SUCCESS {
		util.Logger().Warnf(nil, "service center service already registered, service id %s", respE.ServiceId)
		util.Logger().Warnf(nil, "service center service already registered, service id %s", respE.ServiceId)
		util.Logger().Warnf(nil, "service center service already registered, service id %s", respE.ServiceId)
		util.Logger().Warnf(nil, "service center service already registered, service id %s", respE.ServiceId)
		util.Logger().Warnf(nil, "service center service already registered, service id %s", respE.ServiceId)
		util.Logger().Warnf(nil, "service center service already registered, service id %s", respE.ServiceId)
		respG, err := core.ServiceAPI.GetOne(ctx, core.GetServiceRequest(respE.ServiceId))
		if respG.Response.Code != pb.Response_SUCCESS {
			util.Logger().Errorf(err, "query service center service info failed, service id %s", respE.ServiceId)
			return fmt.Errorf("service center service file lost.")
		}
		core.Service = respG.Service
		return nil
	}
	respS, err := core.ServiceAPI.Create(ctx, core.CreateServiceRequest())
	if err != nil {
		util.Logger().Error("register service center failed", err)
		return err
	}
	core.Service.ServiceId = respS.ServiceId
	util.Logger().Infof("register service center service successfully, service id %s", respS.ServiceId)
	return nil
}

func (s *APIServer) registryInstance(pCtx context.Context) error {
	core.Instance.ServiceId = core.Service.ServiceId

	endpoints := make([]string, 0, len(s.Endpoints))
	for _, address := range s.Endpoints {
		endpoints = append(endpoints, address)
	}

	ctx := core.AddDefaultContextValue(pCtx)
	respI, err := core.InstanceAPI.Register(ctx,
		core.RegisterInstanceRequest(s.HostName, endpoints))
	if respI.GetResponse().Code != pb.Response_SUCCESS {
		err = fmt.Errorf("register service center instance failed, %s", respI.GetResponse().Message)
		util.Logger().Error(err.Error(), nil)
		return err
	}
	core.Instance.InstanceId = respI.InstanceId
	util.Logger().Infof("register service center instance successfully, instance %s/%s, endpoints %s",
		core.Service.ServiceId, respI.InstanceId, endpoints)
	return nil
}

func (s *APIServer) unregisterInstance(pCtx context.Context) error {
	if len(core.Instance.InstanceId) == 0 {
		return nil
	}
	ctx := core.AddDefaultContextValue(pCtx)
	respI, err := core.InstanceAPI.Unregister(ctx, core.UnregisterInstanceRequest())
	if respI.GetResponse().Code != pb.Response_SUCCESS {
		err = fmt.Errorf("unregister service center instance failed, %s", respI.GetResponse().Message)
		util.Logger().Error(err.Error(), nil)
		return err
	}
	util.Logger().Warnf(nil, "unregister service center instance successfully, %s/%s",
		core.Service.ServiceId, core.Instance.InstanceId)
	return nil
}

func (s *APIServer) doAPIServerHeartBeat(pCtx context.Context) {
	if s.isClose {
		return
	}
	ctx := core.AddDefaultContextValue(pCtx)
	respI, err := core.InstanceAPI.Heartbeat(ctx, core.HeartbeatRequest())
	if respI.GetResponse().Code == pb.Response_SUCCESS {
		util.Logger().Debugf("update service center %s heartbeat %s successfully",
			core.Instance.ServiceId, core.Instance.InstanceId)
		return
	}
	util.Logger().Errorf(err, "update service center %s instance %s heartbeat failed",
		core.Instance.ServiceId, core.Instance.InstanceId)

	//服务不存在，创建服务
	err = s.registerServiceCenter()
	if err != nil {
		util.Logger().Errorf(err, "retry to register %s/%s/%s failed.",
			core.Service.AppId, core.Service.ServiceName, core.Service.Version)
	}
}

func (s *APIServer) startHeartBeatService() {
	go func() {
		for {
			select {
			case <-s.err:
				return
			case <-time.After(time.Duration(core.Instance.HealthCheck.Interval) * time.Second):
				s.doAPIServerHeartBeat(context.Background())
			}
		}
	}()
}

func (s *APIServer) graceDone() {
	grace.Before(s.MarkForked)
	grace.After(s.Stop)
	if err := grace.Done(); err != nil {
		util.Logger().Errorf(err, "server reload failed")
	}
}

func (s *APIServer) MarkForked() {
	s.forked = true
}

func (s *APIServer) startRESTServer() (err error) {
	ep, ok := s.Endpoints[REST]
	if !ok {
		return
	}
	s.restSrv, err = rs.NewServer(ep)
	if err != nil {
		return
	}
	util.Logger().Infof("Local listen address: %s, host: %s.", ep, s.HostName)

	go func() {
		err := s.restSrv.Serve()
		if s.isClose {
			return
		}
		util.Logger().Errorf(err, "error to start REST API server %s", ep)
		s.err <- err
	}()
	return
}

func (s *APIServer) startRPCServer() (err error) {
	ep, ok := s.Endpoints[RPC]
	if !ok {
		return
	}
	s.rpcSrv, err = rpc.NewServer(ep)
	if err != nil {
		return
	}
	util.Logger().Infof("Local listen address: %s, host: %s.", ep, s.HostName)

	go func() {
		err := s.rpcSrv.Serve()
		if s.isClose {
			return
		}
		util.Logger().Errorf(err, "error to start RPC API server %s", ep)
		s.err <- err
	}()
	return
}

// 需保证ETCD启动成功后才执行该方法
func (s *APIServer) Start() {
	if !s.isClose {
		return
	}
	s.isClose = false

	var err error
	err = s.startRESTServer()
	if err != nil {
		s.err <- err
		return
	}

	err = s.startRPCServer()
	if err != nil {
		s.err <- err
		return
	}

	s.graceDone()

	// 自注册
	err = s.registerServiceCenter()
	if err != nil {
		s.err <- err
		return
	}

	// 心跳
	s.startHeartBeatService()

	util.Logger().Info("api server is ready")
}

func (s *APIServer) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true

	if !s.forked {
		ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
		s.unregisterInstance(ctx)
	}

	if s.restSrv != nil {
		s.restSrv.Shutdown()
	}

	if s.rpcSrv != nil {
		s.rpcSrv.GracefulStop()
	}

	close(s.err)

	util.Logger().Info("api server stopped.")
}

func GetAPIServer() *APIServer {
	return apiServer
}

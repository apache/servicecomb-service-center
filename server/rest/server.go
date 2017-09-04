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
package rest

import (
	"fmt"
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/util/grace"
	"github.com/ServiceComb/service-center/util/rest"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"net/url"
	"time"
)

var apiServer *APIServer

func init() {
	apiServer = &APIServer{
		isClose: true,
		err:     make(chan error, 1),
	}
}

type APIType int64

func (t APIType) String() string {
	switch t {
	case GRPC:
		return "grpc"
	case REST:
		return "rest"
	default:
		return fmt.Sprintf("SCHEME%d", t)
	}
}

type APIServerConfig struct {
	SSL          bool
	VerifyClient bool
	HostName     string
	Endpoints    map[APIType]string
}

type APIServer struct {
	Config     APIServerConfig
	restIpPort string
	grpcIpPort string
	restSrv    *rest.Server
	grpcSrv    *grpc.Server
	isClose    bool
	forked     bool
	err        chan error
}

const (
	GRPC APIType = 0
	REST APIType = 1
)

func (s *APIServer) Err() <-chan error {
	return s.err
}

func (s *APIServer) parseEndpoint(ep string) (string, error) {
	u, err := url.Parse(ep)
	if err != nil {
		return "", err
	}
	port := u.Port()
	if len(port) > 0 {
		return u.Hostname() + ":" + port, nil
	}
	return u.Hostname(), nil
}

func (s *APIServer) startGrpcServer() {
	var err error

	ep, ok := s.Config.Endpoints[GRPC]
	if !ok {
		return
	}
	ipAddr, err := s.parseEndpoint(ep)
	if err != nil {
		s.err <- err
		return
	}

	s.grpcIpPort = ipAddr

	if s.Config.SSL {
		tlsConfig, err := rest.GetServerTLSConfig(s.Config.VerifyClient)
		if err != nil {
			util.Logger().Error("error to get server tls config", err)
			s.err <- err
			return
		}
		creds := credentials.NewTLS(tlsConfig)
		s.grpcSrv = grpc.NewServer(grpc.Creds(creds))
	} else {
		s.grpcSrv = grpc.NewServer()
	}

	pb.RegisterServiceCtrlServer(s.grpcSrv, ServiceAPI)
	pb.RegisterServiceInstanceCtrlServer(s.grpcSrv, InstanceAPI)

	util.Logger().Infof("listen on server %s", ipAddr)
	ls, err := net.Listen("tcp", ipAddr)
	if err != nil {
		util.Logger().Error("error to start Grpc API server "+ipAddr, err)
		s.err <- err
		return
	}

	go func() {
		err := s.grpcSrv.Serve(ls)
		if s.isClose {
			return
		}

		if s.isClose {
			return
		}

		util.Logger().Errorf(err, "error to start Grpc API server %s", s.grpcIpPort)
		s.err <- err
	}()
}

func (s *APIServer) startRESTfulServer() {
	var err error
	defer func() {
		if err != nil {
			s.err <- err
		}
	}()

	ep, ok := s.Config.Endpoints[REST]
	if !ok {
		return
	}
	ipAddr, err := s.parseEndpoint(ep)
	if err != nil {
		return
	}

	s.restIpPort = ipAddr
	s.restSrv, err = rest.NewServer(ipAddr, nil)
	if err != nil {
		return
	}

	if s.Config.SSL {
		err = s.restSrv.ListenTLS()
	} else {
		err = s.restSrv.Listen()
	}
	if err != nil {
		return
	}

	err = grace.Done()
	if err != nil {
		return
	}

	go func() {
		err := s.restSrv.Serve()
		if s.isClose {
			return
		}
		util.Logger().Errorf(err, "error to start RESTful API server %s", s.restIpPort)
		s.err <- err
	}()
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
	respE, err := ServiceAPI.Exist(ctx, core.GetExistenceRequest())
	if err != nil {
		util.Logger().Error("query service center existence failed", err)
		return err
	}
	if respE.Response.Code == pb.Response_SUCCESS {
		util.Logger().Warnf(nil, "service center service already registered, service id %s", respE.ServiceId)
		respG, err := ServiceAPI.GetOne(ctx, core.GetServiceRequest(respE.ServiceId))
		if respG.Response.Code != pb.Response_SUCCESS {
			util.Logger().Errorf(err, "query service center service info failed, service id %s", respE.ServiceId)
			return fmt.Errorf("service center service file lost.")
		}
		core.Service = respG.Service
		return nil
	}
	respS, err := ServiceAPI.Create(ctx, core.CreateServiceRequest())
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

	endpoints := make([]string, 0, len(s.Config.Endpoints))
	for _, address := range s.Config.Endpoints {
		endpoints = append(endpoints, address)
	}

	ctx := core.AddDefaultContextValue(pCtx)
	respI, err := InstanceAPI.Register(ctx,
		core.RegisterInstanceRequest(s.Config.HostName, endpoints))
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
	respI, err := InstanceAPI.Unregister(ctx, core.UnregisterInstanceRequest())
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
	respI, err := InstanceAPI.Heartbeat(ctx, core.HeartbeatRequest())
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

func (s *APIServer) registerSignalHooks() {
	grace.InitGrace()
	grace.Before(s.registerServerFile)
	grace.After(s.Stop)
}

func (s *APIServer) registerServerFile() {
	s.forked = true
	grace.RegisterFiles(s.restIpPort, s.restSrv.File())
}

// 需保证ETCD启动成功后才执行该方法
func (s *APIServer) Start() {
	if !s.isClose {
		return
	}
	s.isClose = false

	s.registerSignalHooks()

	s.startRESTfulServer()

	s.startGrpcServer()

	// 自注册
	if !grace.IsFork() {
		err := s.registerServiceCenter()
		if err != nil {
			s.err <- err
		}
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

	if s.grpcSrv != nil {
		s.grpcSrv.GracefulStop()
	}

	close(s.err)

	util.Logger().Info("api server stopped.")
}

func GetAPIServer() *APIServer {
	return apiServer
}

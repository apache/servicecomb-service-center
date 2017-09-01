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
package api

import (
	"fmt"
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	rs "github.com/ServiceComb/service-center/server/rest"
	"github.com/ServiceComb/service-center/server/rest/handlers"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/util/rest"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"net/http"
	"net/url"
	"time"
)

type APIType int64

type APIServerConfig struct {
	SSL          bool
	VerifyClient bool
	HostName     string
	Endpoints    map[APIType]string
}

type APIService struct {
	Config APIServerConfig

	grpcSvr *grpc.Server
	isClose bool
	err     chan error
}

const (
	GRPC APIType = 0
	REST APIType = 1
)

func (s *APIService) Err() <-chan error {
	return s.err
}

func (s *APIService) parseEndpoint(ep string) (string, error) {
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

func (s *APIService) startGrpcServer() {
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

	if s.Config.SSL {
		tlsConfig, err := rest.GetServerTLSConfig(s.Config.VerifyClient)
		if err != nil {
			util.Logger().Error("error to get server tls config", err)
			s.err <- err
			return
		}
		creds := credentials.NewTLS(tlsConfig)
		s.grpcSvr = grpc.NewServer(grpc.Creds(creds))
	} else {
		s.grpcSvr = grpc.NewServer()
	}

	pb.RegisterServiceCtrlServer(s.grpcSvr, rs.ServiceAPI)
	pb.RegisterServiceInstanceCtrlServer(s.grpcSvr, rs.InstanceAPI)

	util.Logger().Infof("listen on server %s", ipAddr)
	ls, err := net.Listen("tcp", ipAddr)
	if err != nil {
		util.Logger().Error("error to start Grpc API server "+ipAddr, err)
		s.err <- err
		return
	}

	go func() {
		err := s.grpcSvr.Serve(ls)
		if !s.isClose {
			util.Logger().Error("error to start Grpc API server "+ipAddr, err)
			s.err <- err
		}
	}()
}

func (s *APIService) startRESTfulServer() {
	var err error

	ep, ok := s.Config.Endpoints[REST]
	if !ok {
		return
	}
	ipAddr, err := s.parseEndpoint(ep)
	if err != nil {
		s.err <- err
		return
	}

	http.Handle("/", handlers.DefaultServerHandler())

	go func() {
		if s.Config.SSL {
			err = rest.ListenAndServeTLS(ipAddr, nil)
		} else {
			err = rest.ListenAndServe(ipAddr, nil)
		}

		if !s.isClose {
			util.Logger().Error("error to start RESTful API server "+ipAddr, err)
			s.err <- err
		}
	}()
}

func (s *APIService) registerServiceCenter() error {
	err := s.registryService()
	if err != nil {
		return err
	}
	// 实例信息
	return s.registryInstance()
}

func (s *APIService) registryService() error {
	ctx := core.AddDefaultContextValue(context.TODO())
	respE, err := rs.ServiceAPI.Exist(ctx, core.GetExistenceRequest())
	if err != nil {
		util.Logger().Error("query service center existence failed", err)
		return err
	}
	if respE.Response.Code == pb.Response_SUCCESS {
		util.Logger().Warnf(nil, "service center service already registered, service id %s", respE.ServiceId)
		respG, err := rs.ServiceAPI.GetOne(ctx, core.GetServiceRequest(respE.ServiceId))
		if respG.Response.Code != pb.Response_SUCCESS {
			util.Logger().Errorf(err, "query service center service info failed, service id %s", respE.ServiceId)
			return fmt.Errorf("service center service file lost.")
		}
		core.Service = respG.Service
		return nil
	}
	respS, err := rs.ServiceAPI.Create(ctx, core.CreateServiceRequest())
	if err != nil {
		util.Logger().Error("register service center failed", err)
		return err
	}
	core.Service.ServiceId = respS.ServiceId
	util.Logger().Infof("register service center service successfully, service id %s", respS.ServiceId)
	return nil
}

func (s *APIService) registryInstance() error {
	core.Instance.ServiceId = core.Service.ServiceId

	endpoints := []string{}
	if address, ok := s.Config.Endpoints[GRPC]; ok {
		endpoints = append(endpoints, util.StringJoin([]string{"grpc", address}, "://"))
	}
	if address, ok := s.Config.Endpoints[REST]; ok {
		endpoints = append(endpoints, util.StringJoin([]string{"rest", address}, "://"))
	}

	ctx := core.AddDefaultContextValue(context.TODO())
	respI, err := rs.InstanceAPI.Register(ctx,
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

func (s *APIService) unregisterInstance() error {
	if len(core.Instance.InstanceId) == 0 {
		return nil
	}
	ctx := core.AddDefaultContextValue(context.TODO())
	respI, err := rs.InstanceAPI.Unregister(ctx, core.UnregisterInstanceRequest())
	if respI.GetResponse().Code != pb.Response_SUCCESS {
		err = fmt.Errorf("unregister service center instance failed, %s", respI.GetResponse().Message)
		util.Logger().Error(err.Error(), nil)
		return err
	}
	util.Logger().Warnf(nil, "unregister service center instance successfully, %s/%s",
		core.Service.ServiceId, core.Instance.InstanceId)
	return nil
}

func (s *APIService) doAPIServerHeartBeat() {
	if s.isClose {
		return
	}
	ctx := core.AddDefaultContextValue(context.TODO())
	respI, err := rs.InstanceAPI.Heartbeat(ctx, core.HeartbeatRequest())
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

func (s *APIService) startHeartBeatService() {
	go func() {
		for {
			select {
			case <-s.err:
				return
			case <-time.After(time.Duration(core.Instance.HealthCheck.Interval) * time.Second):
				s.doAPIServerHeartBeat()
			}
		}
	}()
}

// 需保证ETCD启动成功后才执行该方法
func (s *APIService) Start() {
	if !s.isClose {
		return
	}
	s.isClose = false

	s.startRESTfulServer()

	s.startGrpcServer()
	// 自注册
	err := s.registerServiceCenter()
	if err != nil {
		s.err <- err
	}
	// 心跳
	s.startHeartBeatService()

	util.Logger().Info("api server is ready")
}

func (s *APIService) Stop() {
	if s.isClose {
		return
	}

	s.unregisterInstance()

	s.isClose = true

	rest.CloseServer()

	if s.grpcSvr != nil {
		s.grpcSvr.GracefulStop()
	}

	close(s.err)

	util.Logger().Info("api server stopped.")
}

var apiServer *APIService

func init() {
	apiServer = &APIService{
		isClose: true,
		err:     make(chan error, 1),
	}
}

func GetAPIService() *APIService {
	return apiServer
}

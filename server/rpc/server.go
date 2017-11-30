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
package rpc

import (
	"github.com/ServiceComb/service-center/pkg/tlsutil"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
)

type Server struct {
	*grpc.Server
	innerListener net.Listener
}

func (srv *Server) Serve() error {
	return srv.Server.Serve(srv.innerListener)
}

func NewServer(ep string) (_ *Server, err error) {
	ipAddr, err := util.ParseEndpoint(ep)
	if err != nil {
		return
	}

	var grpcSrv *grpc.Server
	if tlsutil.GetServerSSLConfig().SSLEnabled {
		tlsConfig, err := tlsutil.GetServerTLSConfig(tlsutil.GetServerSSLConfig().VerifyClient)
		if err != nil {
			util.Logger().Error("error to get server tls config", err)
			return nil, err
		}
		creds := credentials.NewTLS(tlsConfig)
		grpcSrv = grpc.NewServer(grpc.Creds(creds))
	} else {
		grpcSrv = grpc.NewServer()
	}

	pb.RegisterServiceCtrlServer(grpcSrv, core.ServiceAPI)
	pb.RegisterServiceInstanceCtrlServer(grpcSrv, core.InstanceAPI)
	pb.RegisterGovernServiceCtrlServer(grpcSrv, core.GovernServiceAPI)

	ls, err := net.Listen("tcp", ipAddr)
	if err != nil {
		util.Logger().Error("error to start Grpc API server "+ipAddr, err)
		return
	}

	return &Server{
		Server:        grpcSrv,
		innerListener: ls,
	}, nil
}

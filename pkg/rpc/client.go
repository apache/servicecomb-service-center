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

package rpc

import (
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
)

var (
	ErrAddrEmpty = errors.New("addr is empty")
)

type Config struct {
	Addrs       []string
	Scheme      string
	ServiceName string
}

func GetPickFirstLbConn(config *Config) (*grpc.ClientConn, error) {
	return getLbConn(config.Addrs, config.Scheme, config.ServiceName, func() []grpc.DialOption {
		return []grpc.DialOption{}
	})
}

func GetRoundRobinLbConn(config *Config) (*grpc.ClientConn, error) {
	return getLbConn(config.Addrs, config.Scheme, config.ServiceName, func() []grpc.DialOption {
		return []grpc.DialOption{
			grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		}
	})
}

func getLbConn(addrs []string, scheme, serviceName string, dialOptions func() []grpc.DialOption) (*grpc.ClientConn, error) {
	if len(addrs) <= 0 {
		return nil, ErrAddrEmpty
	}

	addr := make([]resolver.Address, 0, len(addrs))
	for _, a := range addrs {
		addr = append(addr, resolver.Address{Addr: a})
	}

	r := manual.NewBuilderWithScheme(scheme)
	r.InitialState(resolver.State{Addresses: addr})

	opinions := dialOptions()
	opinions = append(opinions, grpc.WithInsecure())
	opinions = append(opinions, grpc.WithResolvers(r))

	conn, err := grpc.Dial(r.Scheme()+":///"+serviceName, opinions...)
	return conn, err
}

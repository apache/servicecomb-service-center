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
	"crypto/tls"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	TLSConfig   *TLSConfig
}

type TLSConfig struct {
	InsecureSkipVerify bool
}

func GetPickFirstLbConn(config *Config) (*grpc.ClientConn, error) {
	return getLbConn(config, func() []grpc.DialOption {
		return []grpc.DialOption{}
	})
}

func GetRoundRobinLbConn(config *Config) (*grpc.ClientConn, error) {
	return getLbConn(config, func() []grpc.DialOption {
		return []grpc.DialOption{
			grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		}
	})
}

func getLbConn(config *Config, dialOptions func() []grpc.DialOption) (*grpc.ClientConn, error) {
	addresses := config.Addrs
	if len(addresses) <= 0 {
		return nil, ErrAddrEmpty
	}

	addr := make([]resolver.Address, 0, len(addresses))
	for _, a := range addresses {
		addr = append(addr, resolver.Address{Addr: a})
	}

	r := manual.NewBuilderWithScheme(config.Scheme)
	r.InitialState(resolver.State{Addresses: addr})

	opinions := dialOptions()
	if config.TLSConfig != nil {
		cfg := &tls.Config{
			InsecureSkipVerify: config.TLSConfig.InsecureSkipVerify,
		}

		opinions = append(opinions, grpc.WithTransportCredentials(credentials.NewTLS(cfg)))
	} else {
		opinions = append(opinions, grpc.WithInsecure())
	}

	opinions = append(opinions, grpc.WithResolvers(r))

	conn, err := grpc.Dial(r.Scheme()+":///"+config.ServiceName, opinions...)
	return conn, err
}

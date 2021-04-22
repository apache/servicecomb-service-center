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

package gov

import (
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
)

const (
	ConfigDistributorKie   = "kie"
	ConfigDistributorIstio = "istio"
	ConfigDistributorMock  = "mock"
)

type NewDistributors func(opts config.DistributorOptions) (ConfigDistributor, error)

var distributors = map[string]ConfigDistributor{}
var distributorPlugins = map[string]NewDistributors{}

//ConfigDistributor persist and distribute Governance policy
//typically, a ConfigDistributor interact with a config server, like ctrip apollo, kie.
//or service mesh system like istio, linkerd.
//ConfigDistributor will convert standard servicecomb gov config to concrete spec, that data plane can recognize.
type ConfigDistributor interface {
	Create(kind, project string, spec []byte) ([]byte, error)
	Update(kind, id, project string, spec []byte) error
	Delete(kind, id, project string) error
	Display(project, app, env string) ([]byte, error)
	List(kind, project, app, env string) ([]byte, error)
	Get(kind, id, project string) ([]byte, error)
	Type() string
	Name() string
}

//InstallDistributor install a plugin to distribute and persist config
func InstallDistributor(t string, newDistributors NewDistributors) {
	distributorPlugins[t] = newDistributors
}

//Init create distributors according to gov config.
//it may creates multiple distributors. and distribute policy one by one
func Init() error {
	distOptions := config.GetGov().DistOptions
	for _, opts := range distOptions {
		if opts.Type == "" {
			log.Warn("empty plugin, skip")
		}
		f, ok := distributorPlugins[opts.Type]
		if !ok {
			log.Warn("unsupported plugin " + opts.Type)
		}
		cd, err := f(opts)
		if err != nil {
			log.Error("can not init config distributor", err)
			return err
		}
		distributors[opts.Name+"::"+opts.Type] = cd
	}
	return nil
}

func Create(kind, project string, spec []byte) ([]byte, error) {
	for _, cd := range distributors {
		return cd.Create(kind, project, spec)
	}
	return nil, nil
}

func List(kind, project, app, env string) ([]byte, error) {
	for _, cd := range distributors {
		return cd.List(kind, project, app, env)
	}
	return nil, nil
}

func Display(project, app, env string) ([]byte, error) {
	for _, cd := range distributors {
		return cd.Display(project, app, env)
	}
	return nil, nil
}

func Get(kind, id, project string) ([]byte, error) {
	for _, cd := range distributors {
		return cd.Get(kind, id, project)
	}
	return nil, nil
}

func Delete(kind, id, project string) error {
	for _, cd := range distributors {
		return cd.Delete(kind, id, project)
	}
	return nil
}

func Update(kind, id, project string, spec []byte) error {
	for _, cd := range distributors {
		return cd.Update(kind, id, project, spec)
	}
	return nil
}

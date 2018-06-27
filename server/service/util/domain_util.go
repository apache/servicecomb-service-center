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
package util

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"strings"
)

func GetAllDomainRawData(ctx context.Context) ([]*backend.KeyValue, error) {
	opts := append(FromContext(ctx),
		registry.WithStrKey(apt.GenerateDomainKey("")),
		registry.WithPrefix())
	rsp, err := backend.Store().Domain().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return rsp.Kvs, nil

}

func GetAllDomain(ctx context.Context) ([]string, error) {
	insWatherByDomainKeys := []string{}
	kvs, err := GetAllDomainRawData(ctx)
	if err != nil {
		return nil, err
	}

	if len(kvs) == 0 {
		return insWatherByDomainKeys, err
	}

	domain := ""
	instByDomain := ""
	arrTmp := []string{}
	for _, kv := range kvs {
		arrTmp = strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
		domain = arrTmp[len(arrTmp)-1]
		instByDomain = apt.GetInstanceRootKey(domain)
		insWatherByDomainKeys = append(insWatherByDomainKeys, instByDomain)
	}
	return insWatherByDomainKeys, err
}

func DomainExist(ctx context.Context, domain string) (bool, error) {
	opts := append(FromContext(ctx),
		registry.WithStrKey(apt.GenerateDomainKey(domain)),
		registry.WithCountOnly())
	rsp, err := backend.Store().Domain().Search(ctx, opts...)
	if err != nil {
		return false, err
	}
	return rsp.Count > 0, nil
}

func ProjectExist(ctx context.Context, domain, project string) (bool, error) {
	opts := append(FromContext(ctx),
		registry.WithStrKey(apt.GenerateProjectKey(domain, project)),
		registry.WithCountOnly())
	rsp, err := backend.Store().Project().Search(ctx, opts...)
	if err != nil {
		return false, err
	}
	return rsp.Count > 0, nil
}

func NewDomain(ctx context.Context, domain string) error {
	_, err := backend.Registry().PutNoOverride(ctx,
		registry.WithStrKey(apt.GenerateDomainKey(domain)))
	if err != nil {
		return err
	}
	return nil
}

func NewProject(ctx context.Context, domain, project string) error {
	_, err := backend.Registry().PutNoOverride(ctx,
		registry.WithStrKey(apt.GenerateProjectKey(domain, project)))
	if err != nil {
		return err
	}
	return nil
}

func NewDomainProject(ctx context.Context, domain, project string) error {
	copyCtx := util.SetContext(util.CloneContext(ctx), CTX_CACHEONLY, "1")
	ok, err := DomainExist(copyCtx, domain)
	if !ok && err == nil {
		err = NewDomain(ctx, domain)
		if err == nil {
			util.Logger().Infof("new domain(%s)", domain)
		}
	}
	if err != nil {
		return err
	}
	ok, err = ProjectExist(copyCtx, domain, project)
	if !ok && err == nil {
		err = NewProject(ctx, domain, project)
		if err == nil {
			util.Logger().Infof("new project(%s/%s)", domain, project)
		}
	}
	return err
}

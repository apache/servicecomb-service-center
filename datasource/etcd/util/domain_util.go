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
	"context"
	"fmt"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/little-cui/etcdadpt"
)

func GetAllDomainRawData(ctx context.Context) ([]*kvstore.KeyValue, error) {
	opts := append(FromContext(ctx),
		etcdadpt.WithStrKey(path.GenerateDomainKey("")),
		etcdadpt.WithPrefix())
	rsp, err := sd.Domain().Search(ctx, opts...)
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
	var arrTmp []string
	for _, keyValue := range kvs {
		arrTmp = strings.Split(util.BytesToStringWithNoCopy(keyValue.Key), "/")
		domain = arrTmp[len(arrTmp)-1]
		instByDomain = path.GetInstanceRootKey(domain)
		insWatherByDomainKeys = append(insWatherByDomainKeys, instByDomain)
	}
	return insWatherByDomainKeys, err
}

func AddDomain(ctx context.Context, domain string) (bool, error) {
	ok, err := etcdadpt.InsertBytes(ctx, path.GenerateDomainKey(domain), nil)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func DomainExist(ctx context.Context, domain string) (bool, error) {
	opts := append(FromContext(ctx),
		etcdadpt.WithStrKey(path.GenerateDomainKey(domain)),
		etcdadpt.WithCountOnly())
	rsp, err := sd.Domain().Search(ctx, opts...)
	if err != nil {
		return false, err
	}
	return rsp.Count > 0, nil
}

func AddProject(ctx context.Context, domain, project string) (bool, error) {
	ok, err := etcdadpt.InsertBytes(ctx, path.GenerateProjectKey(domain, project), nil)
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func ProjectExist(ctx context.Context, domain, project string) (bool, error) {
	opts := append(FromContext(ctx),
		etcdadpt.WithStrKey(path.GenerateProjectKey(domain, project)),
		etcdadpt.WithCountOnly())
	rsp, err := sd.Project().Search(ctx, opts...)
	if err != nil {
		return false, err
	}
	return rsp.Count > 0, nil
}

func NewDomainProject(ctx context.Context, domain, project string) error {
	copyCtx := util.WithCacheOnly(util.CloneContext(ctx))
	ok, err := DomainExist(copyCtx, domain)
	if !ok && err == nil {
		ok, err = AddDomain(ctx, domain)
		if ok {
			log.Info(fmt.Sprintf("new domain(%s)", domain))
		}
	}
	if err != nil {
		return err
	}
	ok, err = ProjectExist(copyCtx, domain, project)
	if !ok && err == nil {
		ok, err = AddProject(ctx, domain, project)
		if ok {
			log.Info(fmt.Sprintf("new project(%s/%s)", domain, project))
		}
	}
	return err
}

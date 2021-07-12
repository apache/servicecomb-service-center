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
	"strings"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func GetAllDomainRawData(ctx context.Context) ([]*sd.KeyValue, error) {
	opts := append(FromContext(ctx),
		client.WithStrKey(path.GenerateDomainKey("")),
		client.WithPrefix())
	rsp, err := kv.Store().Domain().Search(ctx, opts...)
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
	ok, err := client.Instance().PutNoOverride(ctx,
		client.WithStrKey(path.GenerateDomainKey(domain)))
	if err != nil {
		return false, err
	}
	return ok, nil
}

func DomainExist(ctx context.Context, domain string) (bool, error) {
	opts := append(FromContext(ctx),
		client.WithStrKey(path.GenerateDomainKey(domain)),
		client.WithCountOnly())
	rsp, err := kv.Store().Domain().Search(ctx, opts...)
	if err != nil {
		return false, err
	}
	return rsp.Count > 0, nil
}

func AddProject(ctx context.Context, domain, project string) (bool, error) {
	ok, err := client.Instance().PutNoOverride(ctx,
		client.WithStrKey(path.GenerateProjectKey(domain, project)))
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func ProjectExist(ctx context.Context, domain, project string) (bool, error) {
	opts := append(FromContext(ctx),
		client.WithStrKey(path.GenerateProjectKey(domain, project)),
		client.WithCountOnly())
	rsp, err := kv.Store().Project().Search(ctx, opts...)
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
			log.Infof("new domain(%s)", domain)
		}
	}
	if err != nil {
		return err
	}
	ok, err = ProjectExist(copyCtx, domain, project)
	if !ok && err == nil {
		ok, err = AddProject(ctx, domain, project)
		if ok {
			log.Infof("new project(%s/%s)", domain, project)
		}
	}
	return err
}

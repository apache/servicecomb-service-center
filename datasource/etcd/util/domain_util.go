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
	"github.com/apache/servicecomb-service-center/datasource"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
)

func GetAllDomainRawData(ctx context.Context) ([]*sd.KeyValue, error) {
	opts := append(FromContext(ctx),
		client.WithStrKey(apt.GenerateDomainKey("")),
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
	for _, kv := range kvs {
		arrTmp = strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
		domain = arrTmp[len(arrTmp)-1]
		instByDomain = apt.GetInstanceRootKey(domain)
		insWatherByDomainKeys = append(insWatherByDomainKeys, instByDomain)
	}
	return insWatherByDomainKeys, err
}

func DomainExist(ctx context.Context, domain string) (bool, error) {
	return datasource.Instance().DomainExist(ctx, domain)
}

func ProjectExist(ctx context.Context, domain, project string) (bool, error) {
	return datasource.Instance().ProjectExist(ctx, domain, project)
}

func NewDomain(ctx context.Context, domain string) (bool, error) {
	return datasource.Instance().AddDomain(ctx, domain)
}

func NewProject(ctx context.Context, domain, project string) (bool, error) {
	return datasource.Instance().AddProject(ctx, domain, project)
}

func NewDomainProject(ctx context.Context, domain, project string) error {
	copyCtx := util.SetContext(util.CloneContext(ctx), util.CtxCacheOnly, "1")
	ok, err := DomainExist(copyCtx, domain)
	if !ok && err == nil {
		ok, err = NewDomain(ctx, domain)
		if ok {
			log.Infof("new domain(%s)", domain)
		}
	}
	if err != nil {
		return err
	}
	ok, err = ProjectExist(copyCtx, domain, project)
	if !ok && err == nil {
		ok, err = NewProject(ctx, domain, project)
		if ok {
			log.Infof("new project(%s/%s)", domain, project)
		}
	}
	return err
}

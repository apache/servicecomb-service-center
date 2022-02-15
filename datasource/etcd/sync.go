/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except request compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to request writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-chassis/cari/discovery"
	crbac "github.com/go-chassis/cari/rbac"
	"github.com/little-cui/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	esync "github.com/apache/servicecomb-service-center/datasource/etcd/sync"
	"github.com/apache/servicecomb-service-center/pkg/log"
	putil "github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
)

const (
	SyncAllKey     = "/cse-sr/sync-all"
	SyncAllLockKey = "/cse-sr/sync-all-lock"

	// one minutes
	defaultLockTime = 60
)

var (
	ErrWithoutDomainProject = errors.New("key without domain and project")
)

type SyncManager struct {
}

// SyncAll will list all services,accounts,roles,schemas,tags,deps and use tasks to store
func (s *SyncManager) SyncAll(ctx context.Context) error {
	enable := config.GetBool("sync.enableOnStart", false)
	if !enable {
		return nil
	}
	ctx = putil.SetContext(ctx, putil.CtxEnableSync, "1")
	exist, err := etcdadpt.Exist(ctx, SyncAllKey)
	if err != nil {
		return err
	}
	if exist {
		log.Info(fmt.Sprintf("%s key already exists, do not need to do tasks", SyncAllKey))
		return datasource.ErrSyncAllKeyExists
	}
	lock, err := etcdadpt.TryLock(SyncAllLockKey, defaultLockTime)
	if err != nil || lock == nil {
		log.Info(fmt.Sprintf("%s lock not acquired", SyncAllLockKey))
		return nil
	}
	defer func(lock *etcdadpt.DLock) {
		err := lock.Unlock()
		if err != nil {
			log.Error(fmt.Sprintf("fail to unlock the %s key", SyncAllLockKey), err)
		}
	}(lock)
	err = syncAllRoles(ctx)
	if err != nil {
		return err
	}
	err = syncAllAccounts(ctx)
	if err != nil {
		return err
	}
	err = syncAllServices(ctx)
	if err != nil {
		return err
	}
	err = syncAllTags(ctx)
	if err != nil {
		return err
	}
	err = syncAllSchemas(ctx)
	if err != nil {
		return err
	}
	err = syncAllDependencies(ctx)
	if err != nil {
		return err
	}
	return etcdadpt.Put(ctx, SyncAllKey, "1")
}

func syncAllAccounts(ctx context.Context) error {
	kvs, _, err := etcdadpt.List(ctx, path.GenerateRBACAccountKey(""))
	if err != nil {
		return err
	}
	syncOpts := make([]etcdadpt.OpOptions, 0)
	putil.SetDomain(ctx, "")
	putil.SetProject(ctx, "")
	for _, v := range kvs {
		a := &crbac.Account{}
		err = json.Unmarshal(v.Value, a)
		if err != nil {
			log.Error("fail to unmarshal account ", err)
			return err
		}
		opt, err := esync.GenCreateOpts(ctx, datasource.ResourceAccount, a)
		if err != nil {
			log.Error("fail to create sync opts", err)
			return err
		}
		syncOpts = append(syncOpts, opt...)
	}
	err = etcdadpt.Txn(ctx, syncOpts)
	if err != nil {
		log.Error("fail to account tasks", err)
	}
	return err
}

func syncAllRoles(ctx context.Context) error {
	kvs, _, err := etcdadpt.List(ctx, path.GenerateRBACRoleKey(""))
	if err != nil {
		return err
	}
	syncOpts := make([]etcdadpt.OpOptions, 0)
	putil.SetDomain(ctx, "")
	putil.SetProject(ctx, "")
	for _, v := range kvs {
		r := &crbac.Role{}
		err = json.Unmarshal(v.Value, r)
		if err != nil {
			log.Error("fail to unmarshal role", err)
			return err
		}
		opt, err := esync.GenCreateOpts(ctx, datasource.ResourceRole, r)
		if err != nil {
			log.Error("fail to create sync opts", err)
			return err
		}
		syncOpts = append(syncOpts, opt...)
	}
	err = etcdadpt.Txn(ctx, syncOpts)
	if err != nil {
		log.Error("fail to role tasks", err)
	}
	return err
}

// syncAllTags func use kv resource task to store tags
func syncAllTags(ctx context.Context) error {
	kvs, _, err := etcdadpt.List(ctx, path.GetServiceTagRootKey(""))
	if err != nil {
		return err
	}
	syncOpts := make([]etcdadpt.OpOptions, 0)
	for _, kv := range kvs {
		domain, project, err := getDomainProject(string(kv.Key), path.GetServiceTagRootKey(""))
		if err != nil {
			log.Error("fail to get domain and project", err)
			return err
		}
		putil.SetDomain(ctx, domain)
		putil.SetProject(ctx, project)
		opts, err := esync.GenCreateOpts(ctx, datasource.ResourceKV, kv.Value,
			esync.WithOpts(map[string]string{"key": string(kv.Key)}))
		if err != nil {
			log.Error("fail to create tag opts", err)
			return err
		}
		syncOpts = append(syncOpts, opts...)
	}
	err = etcdadpt.Txn(ctx, syncOpts)
	if err != nil {
		log.Error("fail to create tag tasks", err)
	}
	return err
}

func syncAllServices(ctx context.Context) error {
	kvs, _, err := etcdadpt.List(ctx, path.GetServiceRootKey(""))
	if err != nil {
		return err
	}
	syncOpts := make([]etcdadpt.OpOptions, 0)
	for _, kv := range kvs {
		service := &discovery.MicroService{}
		err := json.Unmarshal(kv.Value, service)
		if err != nil {
			log.Error("fail to unmarshal service", err)
			return err
		}
		domain, project, err := getDomainProject(string(kv.Key), path.GetServiceRootKey(""))
		if err != nil {
			log.Error("fail to get domain and project", err)
			return err
		}
		putil.SetDomain(ctx, domain)
		putil.SetProject(ctx, project)
		request := &discovery.CreateServiceRequest{
			Service: service,
		}
		opts, err := esync.GenCreateOpts(ctx, datasource.ResourceService, request)
		if err != nil {
			log.Error("fail to create service task", err)
			return err
		}
		syncOpts = append(syncOpts, opts...)
	}
	err = etcdadpt.Txn(ctx, syncOpts)
	if err != nil {
		log.Error("fail to create service tasks", err)
	}
	return err
}

// syncAllSchemas func use kv resource task to store schemas
func syncAllSchemas(ctx context.Context) error {
	putil.SetDomain(ctx, "")
	putil.SetProject(ctx, "")
	err := syncAllServiceSchemas(ctx)
	if err != nil {
		return err
	}
	err = syncAllServiceSchemaRefs(ctx)
	if err != nil {
		return err
	}
	err = syncAllServiceSchemaContents(ctx)
	if err != nil {
		return err
	}
	return syncAllServiceSchemaSummaries(ctx)
}

func syncAllServiceSchemas(ctx context.Context) error {
	kvs, _, err := etcdadpt.List(ctx, path.GetServiceSchemaRootKey(""))
	if err != nil {
		return err
	}
	syncOpts := make([]etcdadpt.OpOptions, 0)
	for _, kv := range kvs {
		domain, project, err := getDomainProject(string(kv.Key), path.GetServiceSchemaRootKey(""))
		if err != nil {
			log.Error("fail to get domain and project", err)
			return err
		}
		putil.SetDomain(ctx, domain)
		putil.SetProject(ctx, project)
		opts, err := esync.GenCreateOpts(ctx, datasource.ResourceKV, kv.Value,
			esync.WithOpts(map[string]string{"key": string(kv.Key)}))
		if err != nil {
			log.Error("fail to create schema opts", err)
			return err
		}
		syncOpts = append(syncOpts, opts...)
	}
	err = etcdadpt.Txn(ctx, syncOpts)
	if err != nil {
		log.Error("fail to create schema tasks", err)
	}
	return err
}

func syncAllServiceSchemaRefs(ctx context.Context) error {
	kvs, _, err := etcdadpt.List(ctx, path.GetServiceSchemaRefRootKey(""))
	if err != nil {
		return err
	}
	syncOpts := make([]etcdadpt.OpOptions, 0)
	for _, kv := range kvs {
		domain, project, err := getDomainProject(string(kv.Key), path.GetServiceSchemaRefRootKey(""))
		if err != nil {
			log.Error("fail to get domain and project", err)
			return err
		}
		putil.SetDomain(ctx, domain)
		putil.SetProject(ctx, project)
		opts, err := esync.GenCreateOpts(ctx, datasource.ResourceKV, kv.Value,
			esync.WithOpts(map[string]string{"key": string(kv.Key)}))
		if err != nil {
			log.Error("fail to create schema ref opts", err)
			return err
		}
		syncOpts = append(syncOpts, opts...)
	}
	err = etcdadpt.Txn(ctx, syncOpts)
	if err != nil {
		log.Error("fail to create schema ref tasks", err)
	}
	return err
}

func syncAllServiceSchemaContents(ctx context.Context) error {
	kvs, _, err := etcdadpt.List(ctx, path.GetServiceSchemaContentRootKey(""))
	if err != nil {
		return err
	}
	syncOpts := make([]etcdadpt.OpOptions, 0)
	for _, kv := range kvs {
		domain, project, err := getDomainProject(string(kv.Key), path.GetServiceSchemaContentRootKey(""))
		if err != nil {
			log.Error("fail to get domain and project", err)
			return err
		}
		putil.SetDomain(ctx, domain)
		putil.SetProject(ctx, project)
		opts, err := esync.GenCreateOpts(ctx, datasource.ResourceKV, kv.Value,
			esync.WithOpts(map[string]string{"key": string(kv.Key)}))
		if err != nil {
			log.Error("fail to create schema content opts", err)
			return err
		}
		syncOpts = append(syncOpts, opts...)
	}
	err = etcdadpt.Txn(ctx, syncOpts)
	if err != nil {
		log.Error("fail to create schema content tasks", err)
	}
	return err
}

func syncAllServiceSchemaSummaries(ctx context.Context) error {
	kvs, _, err := etcdadpt.List(ctx, path.GetServiceSchemaSummaryRootKey(""))
	if err != nil {
		return err
	}
	syncOpts := make([]etcdadpt.OpOptions, 0)
	for _, kv := range kvs {
		domain, project, err := getDomainProject(string(kv.Key), path.GetServiceSchemaSummaryRootKey(""))
		if err != nil {
			log.Error("fail to get domain and project", err)
			return err
		}
		putil.SetDomain(ctx, domain)
		putil.SetProject(ctx, project)
		opts, err := esync.GenCreateOpts(ctx, datasource.ResourceKV, kv.Value,
			esync.WithOpts(map[string]string{"key": string(kv.Key)}))
		if err != nil {
			log.Error("fail to create schema summary opts", err)
			return err
		}
		syncOpts = append(syncOpts, opts...)
	}
	err = etcdadpt.Txn(ctx, syncOpts)
	if err != nil {
		log.Error("fail to create schema summary tasks", err)
	}
	return err
}

func syncAllDependencies(ctx context.Context) error {
	kvs, _, err := etcdadpt.List(ctx, path.GetServiceDependencyQueueRootKey(""))
	if err != nil {
		return err
	}
	syncOpts := make([]etcdadpt.OpOptions, 0)
	for _, kv := range kvs {
		domain, project, err := getDomainProject(string(kv.Key), path.GetServiceDependencyQueueRootKey(""))
		if err != nil {
			log.Error("fail to get domain and project", err)
			return err
		}
		putil.SetDomain(ctx, domain)
		putil.SetProject(ctx, project)
		opts, err := esync.GenUpdateOpts(ctx, datasource.ResourceKV, kv.Value, esync.WithOpts(map[string]string{"key": string(kv.Key)}))
		if err != nil {
			log.Error("fail to create dep opts", err)
			return err
		}
		syncOpts = append(syncOpts, opts...)
	}
	err = etcdadpt.Txn(ctx, syncOpts)
	if err != nil {
		log.Error("fail to create dep tasks", err)
	}
	return err
}

func getDomainProject(key string, prefixKey string) (domain string, project string, err error) {
	splitKey := strings.Split(key, prefixKey)
	if len(splitKey) != 2 {
		return "", "", ErrWithoutDomainProject
	}
	suffixKey := splitKey[len(splitKey)-1]
	splitStr := strings.Split(suffixKey, "/")
	if len(splitStr) < 2 {
		return "", "", ErrWithoutDomainProject
	}
	domain = splitStr[0]
	project = splitStr[1]
	return
}

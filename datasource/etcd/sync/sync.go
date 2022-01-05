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

package sync

import (
	"context"
	"encoding/json"

	"github.com/go-chassis/cari/sync"
	"github.com/little-cui/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/etcd/key"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type Options struct {
	ResourceID string
	Opts       map[string]string
}

type Option func(options *Options)

func NewSyncOptions() Options {
	return Options{}
}

func WithResourceID(resourceID string) Option {
	return func(options *Options) {
		options.ResourceID = resourceID
	}
}

func WithOpts(opts map[string]string) Option {
	return func(options *Options) {
		options.Opts = opts
	}
}

func GenCreateOpts(ctx context.Context, resourceType string, resource interface{},
	options ...Option) ([]etcdadpt.OpOptions, error) {
	return genOpts(ctx, sync.CreateAction, resourceType, resource, options...)
}

func GenUpdateOpts(ctx context.Context, resourceType string, resource interface{},
	options ...Option) ([]etcdadpt.OpOptions, error) {
	return genOpts(ctx, sync.UpdateAction, resourceType, resource, options...)
}

func GenDeleteOpts(ctx context.Context, resourceType, resourceID string, resource interface{},
	options ...Option) ([]etcdadpt.OpOptions, error) {
	options = append(options, WithResourceID(resourceID))
	return genOpts(ctx, sync.DeleteAction, resourceType, resource, options...)

}

func genOpts(ctx context.Context, action string, resourceType string, resource interface{},
	options ...Option) ([]etcdadpt.OpOptions, error) {
	if !datasource.EnableSync {
		return nil, nil
	}
	syncOpts := NewSyncOptions()
	for _, option := range options {
		option(&syncOpts)
	}
	taskOpt, err := genTaskOpt(ctx, action, resourceType, resource, &syncOpts)
	if err != nil {
		return nil, err
	}
	if action != sync.DeleteAction {
		return []etcdadpt.OpOptions{taskOpt}, nil
	}
	tombstoneOpt, err := genTombstoneOpt(ctx, resourceType, syncOpts.ResourceID)
	if err != nil {
		return nil, err
	}
	return []etcdadpt.OpOptions{taskOpt, tombstoneOpt}, nil
}

func genTaskOpt(ctx context.Context, action string, resourceType string, resource interface{},
	syncOpts *Options) (etcdadpt.OpOptions, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	task, err := sync.NewTask(domain, project, action, resourceType, resource)
	if err != nil {
		return etcdadpt.OpOptions{}, err
	}
	if syncOpts.Opts != nil {
		task.Opts = syncOpts.Opts
	}
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return etcdadpt.OpOptions{}, err
	}
	taskOpPut := etcdadpt.OpPut(etcdadpt.WithStrKey(key.TaskKey(domain, project,
		task.ID, task.Timestamp)), etcdadpt.WithValue(taskBytes))
	return taskOpPut, nil
}

func genTombstoneOpt(ctx context.Context, resourceType, resourceID string) (etcdadpt.OpOptions, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	tombstone := sync.NewTombstone(domain, project, resourceType, resourceID)
	tombstoneBytes, err := json.Marshal(tombstone)
	if err != nil {
		return etcdadpt.OpOptions{}, err
	}
	tombstoneOpPut := etcdadpt.OpPut(etcdadpt.WithStrKey(key.TombstoneKey(domain, project, tombstone.ResourceType,
		tombstone.ResourceID)), etcdadpt.WithValue(tombstoneBytes))
	return tombstoneOpPut, nil
}

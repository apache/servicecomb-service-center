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

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/cari/sync"
)

const (
	CollectionTask      = "task"
	CollectionTombstone = "tombstone"
)

type Options struct {
	ResourceID string
	Opts       map[string]string
}

type Option func(options *Options)

func NewSyncOption() Options {
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

func DoCreateOpts(ctx context.Context, resourceType string, resource interface{}, options ...Option) error {
	return doOpts(ctx, sync.CreateAction, resourceType, resource, options...)
}

func DoUpdateOpts(ctx context.Context, resourceType string, resource interface{}, options ...Option) error {
	return doOpts(ctx, sync.UpdateAction, resourceType, resource, options...)
}

func DoDeleteOpts(ctx context.Context, resourceType, resourceID string, resource interface{}, options ...Option) error {
	options = append(options, WithResourceID(resourceID))
	return doOpts(ctx, sync.DeleteAction, resourceType, resource, options...)
}

func doOpts(ctx context.Context, action string, resourceType string, resource interface{}, options ...Option) error {
	if !util.EnableSync(ctx) {
		return nil
	}
	syncOpts := NewSyncOption()
	for _, option := range options {
		option(&syncOpts)
	}
	err := doTaskOpt(ctx, action, resourceType, resource, &syncOpts)
	if err != nil || action != sync.DeleteAction {
		return err
	}
	return doTombstoneOpt(ctx, resourceType, syncOpts.ResourceID)
}

func doTaskOpt(ctx context.Context, action string, resourceType string, resource interface{}, syncOpts *Options) error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	task, err := sync.NewTask(domain, project, action, resourceType, resource)
	if err != nil {
		return err
	}
	if syncOpts != nil {
		task.Opts = syncOpts.Opts
	}
	_, err = mongo.GetClient().GetDB().Collection(CollectionTask).InsertOne(ctx, task)
	return err
}

func doTombstoneOpt(ctx context.Context, resourceType, resourceID string) error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	tombstone := sync.NewTombstone(domain, project, resourceType, resourceID)
	_, err := mongo.GetClient().GetDB().Collection(CollectionTombstone).InsertOne(ctx, tombstone)
	return err
}

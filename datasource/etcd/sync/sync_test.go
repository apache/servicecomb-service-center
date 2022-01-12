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

package sync_test

import (
	"context"
	"testing"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sync"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func optsContext() context.Context {
	ctx := util.WithNoCache(util.SetDomainProject(context.Background(), "sync-opts",
		"sync-opts"))
	return util.WithNoCache(util.SetContext(ctx, util.CtxEnableSync, "1"))
}

func TestOpts(t *testing.T) {
	t.Run("create func will create a task opt should pass", func(t *testing.T) {
		opts, err := sync.GenCreateOpts(optsContext(), datasource.ResourceService, &pb.CreateServiceRequest{})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(opts))
	})

	t.Run("update func will create a task opt should pass", func(t *testing.T) {
		opts, err := sync.GenUpdateOpts(optsContext(), datasource.ResourceService, &pb.UpdateServicePropsRequest{})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(opts))
	})

	t.Run("delete func will create a task and a tombstone should pass", func(t *testing.T) {
		opts, err := sync.GenDeleteOpts(optsContext(), datasource.ResourceService, "11111", &pb.DeleteServiceRequest{})
		assert.Nil(t, err)
		assert.Equal(t, 2, len(opts))
	})
}

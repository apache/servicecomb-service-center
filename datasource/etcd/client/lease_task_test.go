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
package client_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-chassis/etcdadpt"
	"github.com/go-chassis/etcdadpt/buildin"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
)

type mockRegistry struct {
	*buildin.Client
	LeaseErr error
}

func (c *mockRegistry) LeaseRenew(ctx context.Context, leaseID int64) (TTL int64, err error) {
	if c.LeaseErr != nil {
		return 0, c.LeaseErr
	}
	return 1, nil
}

func TestLeaseTask_Do(t *testing.T) {
	c := &mockRegistry{}
	lt := client.NewLeaseAsyncTask(etcdadpt.OptionsToOp(etcdadpt.WithStrKey("/a"), etcdadpt.WithLease(1)))
	lt.Client = c

	c.LeaseErr = errors.New("other error")
	err := lt.Do(context.Background())
	assert.NoError(t, err)

	c.LeaseErr = etcdadpt.ErrLeaseNotFound
	err = lt.Do(context.Background())
	assert.Error(t, err)
}

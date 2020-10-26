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
	"fmt"
	. "github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client/buildin"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"testing"
)

type mockRegistry struct {
	*buildin.Registry
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
	lt := NewLeaseAsyncTask(OptionsToOp(WithStrKey("/a"), WithLease(1)))
	lt.Client = c

	c.LeaseErr = errorsEx.InternalError("lease not found")
	err := lt.Do(context.Background())
	if err != nil || lt.Err() != nil {
		t.Fatalf("TestLeaseTask_Do failed")
	}

	c.LeaseErr = fmt.Errorf("network error")
	err = lt.Do(context.Background())
	if err == nil || lt.Err() == nil {
		t.Fatalf("TestLeaseTask_Do failed")
	}
}

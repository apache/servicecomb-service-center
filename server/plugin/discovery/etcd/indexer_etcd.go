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
package etcd

import (
	"fmt"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"

	"context"
)

// Indexer implements discovery.Indexer.
// Indexer searches data from etcd server.
type Indexer struct {
	Client registry.Registry
	Parser proto.Parser
	Root   string
}

func (i *Indexer) CheckPrefix(key string) error {
	if strings.Index(key, i.Root) != 0 {
		return fmt.Errorf("search '%s' mismatch pattern %s", key, i.Root)
	}
	return nil
}

func (i *Indexer) Search(ctx context.Context, opts ...registry.PluginOpOption) (*discovery.Response, error) {
	op := registry.OpGet(opts...)
	key := util.BytesToStringWithNoCopy(op.Key)

	log.Debugf("search '%s' match special options, request etcd server, opts: %s", key, op)

	if err := i.CheckPrefix(key); err != nil {
		return nil, err
	}

	resp, err := i.Client.Do(ctx, opts...)
	if err != nil {
		return nil, err
	}

	r := new(discovery.Response)
	r.Count = resp.Count
	if len(resp.Kvs) == 0 || op.CountOnly {
		return r, nil
	}

	p := i.Parser
	if op.KeyOnly {
		p = nil
	}

	kvs := make([]*discovery.KeyValue, 0, len(resp.Kvs))
	for _, src := range resp.Kvs {
		kv := discovery.NewKeyValue()
		if err := FromEtcdKeyValue(kv, src, p); err != nil {
			continue
		}
		kvs = append(kvs, kv)
	}
	r.Kvs = kvs
	return r, nil
}

// Creditable implements discovery.Indexer.Creditable.
func (i *Indexer) Creditable() bool {
	return true
}

func NewEtcdIndexer(root string, p proto.Parser) (indexer *Indexer) {
	return &Indexer{Client: backend.Registry(), Parser: p, Root: root}
}

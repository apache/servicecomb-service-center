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
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"

	"context"
)

// Indexer implements sd.Indexer.
// Indexer searches data from etcd server.
type Indexer struct {
	Client client.Registry
	Parser proto.Parser
	Root   string
}

func (i *Indexer) CheckPrefix(key string) error {
	if strings.Index(key, i.Root) != 0 {
		return fmt.Errorf("search '%s' mismatch pattern %s", key, i.Root)
	}
	return nil
}

func (i *Indexer) Search(ctx context.Context, opts ...client.PluginOpOption) (r *sd.Response, err error) {
	op := client.OpGet(opts...)
	key := util.BytesToStringWithNoCopy(op.Key)

	log.Debugf("search '%s' match special options, request etcd server, opts: %s", key, op)

	if err := i.CheckPrefix(key); err != nil {
		return nil, err
	}

	resp, err := i.Client.Do(ctx, opts...)
	if err != nil {
		return nil, err
	}

	r = new(sd.Response)
	r.Count = resp.Count
	if len(resp.Kvs) == 0 || op.CountOnly {
		return
	}

	p := i.Parser
	if op.KeyOnly {
		p = nil
	}

	kvs := make([]*sd.KeyValue, 0, len(resp.Kvs))
	for _, src := range resp.Kvs {
		kv := sd.NewKeyValue()
		if err = FromEtcdKeyValue(kv, src, p); err != nil {
			continue
		}
		kvs = append(kvs, kv)
	}
	r.Kvs = kvs
	return
}

// Creditable implements sd.Indexer#Creditable.
func (i *Indexer) Creditable() bool {
	return true
}

func NewEtcdIndexer(root string, p proto.Parser) (indexer *Indexer) {
	return &Indexer{Client: kv.Registry(), Parser: p, Root: root}
}

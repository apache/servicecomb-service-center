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
package backend

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"strings"
)

type CommonIndexer struct {
	Client registry.Registry
	Parser Parser
	Root   string
}

func (i *CommonIndexer) CheckPrefix(key string) error {
	if strings.Index(key, i.Root) != 0 {
		return fmt.Errorf("search '%s' mismatch pattern %s", key, i.Root)
	}
	return nil
}

func (i *CommonIndexer) Search(ctx context.Context, opts ...registry.PluginOpOption) (r *Response, err error) {
	op := registry.OpGet(opts...)

	key := util.BytesToStringWithNoCopy(op.Key)

	if err := i.CheckPrefix(key); err != nil {
		return nil, err
	}

	resp, err := i.Client.Do(ctx, opts...)
	if err != nil {
		return nil, err
	}

	r = new(Response)
	r.Count = resp.Count
	if len(resp.Kvs) == 0 || op.CountOnly {
		return
	}

	p := i.Parser
	if op.KeyOnly {
		p = nil
	}

	kvs := make([]*KeyValue, 0, len(resp.Kvs))
	for _, src := range resp.Kvs {
		kv := new(KeyValue)
		if err = kv.From(p, src); err != nil {
			continue
		}
		kvs = append(kvs, kv)
	}
	r.Kvs = kvs
	return
}

func NewCommonIndexer(root string, p Parser) (indexer *CommonIndexer) {
	return &CommonIndexer{Client: Registry(), Parser: p, Root: root}
}

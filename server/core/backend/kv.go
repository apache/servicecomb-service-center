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
	"github.com/coreos/etcd/mvcc/mvccpb"
)

type KeyValue struct {
	Key            []byte
	Value          interface{}
	Version        int64
	CreateRevision int64
	ModRevision    int64
}

func (kv *KeyValue) From(p *Parser, s *mvccpb.KeyValue) (err error) {
	kv.Key = s.Key
	kv.Version = s.Version
	kv.CreateRevision = s.CreateRevision
	kv.ModRevision = s.ModRevision
	if p == nil {
		return
	}
	kv.Value, err = p.Unmarshal(s.Value)
	return
}

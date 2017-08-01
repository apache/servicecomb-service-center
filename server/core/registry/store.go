//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package registry

import (
	"github.com/coreos/etcd/mvcc/mvccpb"
	"sync"
)

type Store interface {
	Version() int64
	Get() (interface{}, bool)
	Set(interface{}, interface{})
}

type KvStore struct {
	kvs    map[string]*mvccpb.KeyValue
	rev    int64
	kvsMux sync.RWMutex
	revMux sync.RWMutex
}

func (s *KvStore) Get(key string) (v *mvccpb.KeyValue, b bool) {
	s.kvsMux.RLock()
	v, b = s.kvs[key]
	s.kvsMux.RUnlock()
	return
}

func (s *KvStore) Set(key string, value *mvccpb.KeyValue) {
	s.kvsMux.Lock()
	if value == nil {
		delete(s.kvs, key)
	} else {
		s.kvs[key] = value
	}
	s.kvsMux.Unlock()
}

func (s *KvStore) Map() map[string]*mvccpb.KeyValue {
	return s.kvs
}

func (s *KvStore) Len() int {
	return len(s.kvs)
}

func NewKvStore() *KvStore {
	return &KvStore{
		kvs: make(map[string]*mvccpb.KeyValue),
	}
}

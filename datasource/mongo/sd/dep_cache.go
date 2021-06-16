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

package sd

import (
	"reflect"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
	"go.mongodb.org/mongo-driver/bson"
)

type depStore struct {
	dirty      bool
	d          *DocStore
	indexCache *IndexCache
}

func init() {
	RegisterCacher(dep, newDepStore)
}

func newDepStore() *MongoCacher {
	options := DefaultOptions().SetTable(dep)
	cache := &depStore{
		dirty:      false,
		d:          NewDocStore(),
		indexCache: NewIndexCache(),
	}
	depUnmarshal := func(doc bson.Raw) (resource sdcommon.Resource) {
		docID := MongoDocument{}
		err := bson.Unmarshal(doc, &docID)
		if err != nil {
			return
		}
		dep := model.DependencyRule{}
		err = bson.Unmarshal(doc, &dep)
		if err != nil {
			return
		}
		resource.Value = dep
		resource.Key = docID.ID.Hex()
		return
	}
	return NewMongoCacher(options, cache, depUnmarshal)
}

func (s *depStore) Name() string {
	return dep
}

func (s *depStore) Size() int {
	return s.d.Size()
}

func (s *depStore) Get(key string) interface{} {
	return s.d.Get(key)
}

func (s *depStore) ForEach(iter func(k string, v interface{}) (next bool)) {
	s.d.ForEach(iter)
}

func (s *depStore) GetValue(index string) []interface{} {
	docs := s.indexCache.Get(index)
	res := make([]interface{}, 0, len(docs))
	for _, v := range docs {
		res = append(res, s.d.Get(v))
	}
	return res
}

func (s *depStore) Dirty() bool {
	return s.dirty
}

func (s *depStore) MarkDirty() {
	s.dirty = true
}

func (s *depStore) Clear() {
	s.dirty = false
	s.d.store.Flush()
}

func (s *depStore) ProcessUpdate(event MongoEvent) {
	dep, ok := event.Value.(model.DependencyRule)
	if !ok {
		return
	}
	if dep.ServiceKey == nil {
		return
	}
	s.d.Put(event.DocumentID, event.Value)
	s.indexCache.Put(genDepServiceKey(dep), event.DocumentID)
}

func (s *depStore) ProcessDelete(event MongoEvent) {
	dep, ok := s.d.Get(event.DocumentID).(model.DependencyRule)
	if !ok {
		return
	}
	if dep.ServiceKey == nil {
		return
	}
	s.d.DeleteDoc(event.DocumentID)
	s.indexCache.Delete(genDepServiceKey(dep), event.DocumentID)
}

func (s *depStore) isValueNotUpdated(value interface{}, newValue interface{}) bool {
	newDep, ok := newValue.(model.DependencyRule)
	if !ok {
		return true
	}
	oldDep, ok := value.(model.DependencyRule)
	if !ok {
		return true
	}
	return reflect.DeepEqual(newDep, oldDep)
}

func genDepServiceKey(dep model.DependencyRule) string {
	return strings.Join([]string{dep.Type, dep.ServiceKey.AppId, dep.ServiceKey.ServiceName, dep.ServiceKey.Version}, "/")
}

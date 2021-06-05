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
	rmodel "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
)

type instanceStore struct {
	dirty      bool
	d          *DocStore
	indexCache *IndexCache
}

func init() {
	RegisterCacher(instance, newInstanceStore)
}

func newInstanceStore() *MongoCacher {
	options := DefaultOptions().SetTable(instance)
	cache := &instanceStore{
		dirty:      false,
		d:          NewDocStore(),
		indexCache: NewIndexCache(),
	}
	instanceUnmarshal := func(doc bson.Raw) (resource sdcommon.Resource) {
		docID := MongoDocument{}
		err := bson.Unmarshal(doc, &docID)
		if err != nil {
			return
		}
		inst := model.Instance{}
		err = bson.Unmarshal(doc, &inst)
		if err != nil {
			return
		}
		resource.Value = inst
		resource.Key = docID.ID.Hex()
		return
	}
	return NewMongoCacher(options, cache, instanceUnmarshal)
}

func (s *instanceStore) Name() string {
	return instance
}

func (s *instanceStore) Size() int {
	return s.d.Size()
}

func (s *instanceStore) Get(key string) interface{} {
	return s.d.Get(key)
}

func (s *instanceStore) ForEach(iter func(k string, v interface{}) (next bool)) {
	s.d.ForEach(iter)
}

func (s *instanceStore) GetValue(index string) []interface{} {
	docs := s.indexCache.Get(index)
	res := make([]interface{}, 0, len(docs))
	for _, v := range docs {
		res = append(res, s.d.Get(v))
	}
	return res
}

func (s *instanceStore) Dirty() bool {
	return s.dirty
}

func (s *instanceStore) MarkDirty() {
	s.dirty = true
}

func (s *instanceStore) Clear() {
	s.dirty = false
	s.d.store.Flush()
}

func (s *instanceStore) ProcessUpdate(event MongoEvent) {
	//instance only process create and del event.
	if event.Type == rmodel.EVT_UPDATE {
		return
	}
	inst, ok := event.Value.(model.Instance)
	if !ok {
		return
	}
	s.d.Put(event.DocumentID, event.Value)
	s.indexCache.Put(genInstServiceID(inst), event.DocumentID)
}

func (s *instanceStore) ProcessDelete(event MongoEvent) {
	inst, ok := s.d.Get(event.DocumentID).(model.Instance)
	if !ok {
		return
	}
	s.d.DeleteDoc(event.DocumentID)
	s.indexCache.Delete(genInstServiceID(inst), event.DocumentID)
}

func (s *instanceStore) isValueNotUpdated(value interface{}, newValue interface{}) bool {
	return true
}

func genInstServiceID(inst model.Instance) string {
	return inst.Instance.ServiceId
}

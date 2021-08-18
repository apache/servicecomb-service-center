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

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
	cmap "github.com/orcaman/concurrent-map"
	"go.mongodb.org/mongo-driver/bson"
)

type instanceStore struct {
	dirty bool
	// the key is documentID, is value is mongo document.
	concurrentMap cmap.ConcurrentMap
	// the key is generated by indexFuncs,the value is a set of documentID.
	indexSets IndexCache
}

func init() {
	RegisterCacher(instance, newInstanceStore)
	InstIndexCols = NewIndexCols()
	InstIndexCols.AddIndexFunc(InstServiceIDIndex)
}

func newInstanceStore() *MongoCacher {
	options := DefaultOptions().SetTable(instance)
	cache := &instanceStore{
		dirty:         false,
		concurrentMap: cmap.New(),
		indexSets:     NewIndexCache(),
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
	return s.concurrentMap.Count()
}

func (s *instanceStore) Get(key string) interface{} {
	if v, exist := s.concurrentMap.Get(key); exist {
		return v
	}
	return nil
}

func (s *instanceStore) ForEach(iter func(k string, v interface{}) (next bool)) {
	for k, v := range s.concurrentMap.Items() {
		if !iter(k, v) {
			break
		}
	}
}

func (s *instanceStore) GetValue(index string) []interface{} {
	docs := s.indexSets.Get(index)
	res := make([]interface{}, 0, len(docs))
	for _, id := range docs {
		if doc, exist := s.concurrentMap.Get(id); exist {
			res = append(res, doc)
		}
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
	s.concurrentMap.Clear()
	s.indexSets.Clear()
}

func (s *instanceStore) ProcessUpdate(event MongoEvent) {
	instData, ok := event.Value.(model.Instance)
	if !ok {
		return
	}
	if instData.Instance == nil {
		return
	}
	// set the document data.
	s.concurrentMap.Set(event.DocumentID, event.Value)
	for _, index := range InstIndexCols.GetIndexes(instData) {
		// set the index sets.
		s.indexSets.Put(index, event.DocumentID)
	}
}

func (s *instanceStore) ProcessDelete(event MongoEvent) {
	instanceData, ok := s.concurrentMap.Get(event.DocumentID)
	if !ok {
		return
	}
	instMongo := instanceData.(model.Instance)
	if instMongo.Instance == nil {
		return
	}
	s.concurrentMap.Remove(event.DocumentID)
	for _, index := range InstIndexCols.GetIndexes(instanceData) {
		s.indexSets.Delete(index, event.DocumentID)
	}
}

func (s *instanceStore) isValueNotUpdated(value interface{}, newValue interface{}) bool {
	newInst, ok := newValue.(model.Instance)
	if !ok {
		return true
	}
	oldInst, ok := value.(model.Instance)
	if !ok {
		return true
	}
	newInst.RefreshTime = oldInst.RefreshTime
	return reflect.DeepEqual(newInst, oldInst)
}

func InstServiceIDIndex(data interface{}) string {
	inst := data.(model.Instance)
	return strings.Join([]string{inst.Domain, inst.Project, inst.Instance.ServiceId}, datasource.SPLIT)
}

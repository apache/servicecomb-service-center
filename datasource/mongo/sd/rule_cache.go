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
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
	"go.mongodb.org/mongo-driver/bson"
)

type ruleStore struct {
	dirty      bool
	d          *DocStore
	indexCache *IndexCache
}

func init() {
	RegisterCacher(rule, newRuleStore)
}

func newRuleStore() *MongoCacher {
	options := DefaultOptions().SetTable(rule)
	cache := &ruleStore{
		dirty:      false,
		d:          NewDocStore(),
		indexCache: NewIndexCache(),
	}
	ruleUnmarshal := func(doc bson.Raw) (resource sdcommon.Resource) {
		docID := MongoDocument{}
		err := bson.Unmarshal(doc, &docID)
		if err != nil {
			return
		}
		rule := model.Rule{}
		err = bson.Unmarshal(doc, &rule)
		if err != nil {
			return
		}
		resource.Value = rule
		resource.Key = docID.ID.Hex()
		return
	}
	return NewMongoCacher(options, cache, ruleUnmarshal)
}

func (s *ruleStore) Name() string {
	return rule
}

func (s *ruleStore) Size() int {
	return s.d.Size()
}

func (s *ruleStore) Get(key string) interface{} {
	return s.d.Get(key)
}

func (s *ruleStore) ForEach(iter func(k string, v interface{}) (next bool)) {
	s.d.ForEach(iter)
}

func (s *ruleStore) GetValue(index string) []interface{} {
	docs := s.indexCache.Get(index)
	res := make([]interface{}, 0, len(docs))
	for _, v := range docs {
		res = append(res, s.d.Get(v))
	}
	return res
}

func (s *ruleStore) Dirty() bool {
	return s.dirty
}

func (s *ruleStore) MarkDirty() {
	s.dirty = true
}

func (s *ruleStore) Clear() {
	s.dirty = false
	s.d.store.Flush()
}

func (s *ruleStore) ProcessUpdate(event MongoEvent) {
	rule, ok := event.Value.(model.Rule)
	if !ok {
		return
	}
	s.d.Put(event.DocumentID, event.Value)
	s.indexCache.Put(genRuleServiceID(rule), event.DocumentID)
}

func (s *ruleStore) ProcessDelete(event MongoEvent) {
	rule, ok := s.d.Get(event.DocumentID).(model.Rule)
	if !ok {
		return
	}
	s.d.DeleteDoc(event.DocumentID)
	s.indexCache.Delete(genRuleServiceID(rule), event.DocumentID)
}

func (s *ruleStore) isValueNotUpdated(value interface{}, newValue interface{}) bool {
	return false
}

func genRuleServiceID(rule model.Rule) string {
	return rule.ServiceID
}

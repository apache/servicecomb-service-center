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
	"strings"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
	"go.mongodb.org/mongo-driver/bson"
)

type serviceStore struct {
	dirty      bool
	docCache   *DocStore
	indexCache *indexCache
}

func init() {
	RegisterCacher(service, newServiceStore)
}

func newServiceStore() *MongoCacher {
	options := DefaultOptions().SetTable(service)
	cache := &serviceStore{
		dirty:      false,
		docCache:   NewDocStore(),
		indexCache: NewIndexCache(),
	}
	serviceUnmarshal := func(doc bson.Raw) (resource sdcommon.Resource) {
		docID := MongoDocument{}
		err := bson.Unmarshal(doc, &docID)
		if err != nil {
			return
		}
		service := model.Service{}
		err = bson.Unmarshal(doc, &service)
		if err != nil {
			return
		}
		resource.Value = service
		resource.Key = docID.ID.Hex()
		return
	}
	return NewMongoCacher(options, cache, serviceUnmarshal)
}

func (s *serviceStore) Name() string {
	return service
}

func (s *serviceStore) Size() int {
	return s.docCache.Size()
}

func (s *serviceStore) Get(key string) interface{} {
	return s.docCache.Get(key)
}

func (s *serviceStore) ForEach(iter func(k string, v interface{}) (next bool)) {
	s.docCache.ForEach(iter)
}

func (s *serviceStore) GetValue(index string) []interface{} {
	var docs []string
	docs = s.indexCache.Get(index)
	res := make([]interface{}, 0, len(docs))
	for _, v := range docs {
		res = append(res, s.docCache.Get(v))
	}
	return res
}

func (s *serviceStore) Dirty() bool {
	return s.dirty
}

func (s *serviceStore) MarkDirty() {
	s.dirty = true
}

func (s *serviceStore) Clear() {
	s.dirty = false
	s.docCache.store.Flush()
}

func (s *serviceStore) ProcessUpdate(event MongoEvent) {
	service, ok := event.Value.(model.Service)
	if !ok {
		return
	}
	s.docCache.Put(event.DocumentID, event.Value)
	s.indexCache.Put(genServiceID(service), event.DocumentID)
	s.indexCache.Put(getServiceInfo(service), event.DocumentID)
	return
}

func (s *serviceStore) ProcessDelete(event MongoEvent) {
	service, ok := s.docCache.Get(event.DocumentID).(model.Service)
	if !ok {
		return
	}
	s.docCache.DeleteDoc(event.DocumentID)
	s.indexCache.Delete(genServiceID(service), event.DocumentID)
	s.indexCache.Delete(getServiceInfo(service), event.DocumentID)
}

func (s *serviceStore) isValueNotUpdated(value interface{}, newValue interface{}) bool {
	return false
}

func genServiceID(svc model.Service) string {
	return svc.Service.ServiceId
}

func getServiceInfo(svc model.Service) string {
	return strings.Join([]string{svc.Domain, svc.Project, svc.Service.AppId, svc.Service.ServiceName, svc.Service.Version}, "/")
}

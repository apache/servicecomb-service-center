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

package disco_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/resource/disco"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func init() {
	rest.RegisterServant(&disco.SchemaResource{})
}

func TestSchemaRouter_PutSchema(t *testing.T) {
	ctx := context.TODO()

	service, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{Service: &pb.MicroService{
		ServiceName: "put_schema",
		Schemas:     []string{"1"},
	}})
	assert.NoError(t, err)
	serviceID := service.ServiceId
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID})

	t.Run("put schema, should ok", func(t *testing.T) {
		requestBody := `{
  "schema": "xxx",
  "summary": "hash"
}`
		r, _ := http.NewRequest(http.MethodPut, "/v4/default/registry/microservices/"+serviceID+"/schemas/1", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("put schema again, should ok", func(t *testing.T) {
		requestBody := `{
  "schema": "yyy"
}`
		r, _ := http.NewRequest(http.MethodPut, "/v4/default/registry/microservices/"+serviceID+"/schemas/1", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("put a new schema, should ok", func(t *testing.T) {
		requestBody := `{
  "schema": "zzz"
}`
		r, _ := http.NewRequest(http.MethodPut, "/v4/default/registry/microservices/"+serviceID+"/schemas/2", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		service, err := discosvc.GetService(util.WithNoCache(ctx), &pb.GetServiceRequest{ServiceId: serviceID})
		assert.NoError(t, err)
		assert.Equal(t, []string{"1", "2"}, service.Schemas)
	})

	t.Run("put an invalid schema, should failed", func(t *testing.T) {
		requestBody := ``
		r, _ := http.NewRequest(http.MethodPut, "/v4/default/registry/microservices/"+serviceID+"/schemas/2", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSchemaRouter_GetSchema(t *testing.T) {
	ctx := context.TODO()

	service, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{Service: &pb.MicroService{
		ServiceName: "get_schema",
		Schemas:     []string{"1"},
	}})
	assert.NoError(t, err)
	serviceID := service.ServiceId
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID})

	t.Run("get a not exist schema, should failed", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/v4/default/registry/microservices/"+serviceID+"/schemas/1", nil)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("get schema, should ok", func(t *testing.T) {
		requestBody := `{
  "schema": "xxx",
  "summary": "hash"
}`
		r, _ := http.NewRequest(http.MethodPut, "/v4/default/registry/microservices/"+serviceID+"/schemas/1", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		r, _ = http.NewRequest(http.MethodGet, "/v4/default/registry/microservices/"+serviceID+"/schemas/1", nil)
		w = httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp pb.GetSchemaResponse
		body, _ := ioutil.ReadAll(w.Body)
		err := json.Unmarshal(body, &resp)
		assert.NoError(t, err)
		assert.Equal(t, "xxx", resp.Schema)
		assert.Empty(t, resp.SchemaSummary)
		assert.Equal(t, "hash", w.Header().Get("X-Schema-Summary"))
	})
}

func TestSchemaRouter_DeleteSchema(t *testing.T) {
	ctx := context.TODO()

	service, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{Service: &pb.MicroService{
		ServiceName: "delete_schema",
		Schemas:     []string{"1"},
	}})
	assert.NoError(t, err)
	serviceID := service.ServiceId
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID})

	t.Run("delete a not exist schema, should failed", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodDelete, "/v4/default/registry/microservices/"+serviceID+"/schemas/1", nil)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("delete schema, should ok", func(t *testing.T) {
		requestBody := `{
  "schema": "xxx",
  "summary": "hash"
}`
		r, _ := http.NewRequest(http.MethodPut, "/v4/default/registry/microservices/"+serviceID+"/schemas/1", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		r, _ = http.NewRequest(http.MethodDelete, "/v4/default/registry/microservices/"+serviceID+"/schemas/1", nil)
		w = httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		r, _ = http.NewRequest(http.MethodGet, "/v4/default/registry/microservices/"+serviceID+"/schemas/1", nil)
		w = httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSchemaRouter_PutSchemas(t *testing.T) {
	ctx := context.TODO()

	service, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{Service: &pb.MicroService{
		ServiceName: "put_schema",
		Schemas:     []string{"1"},
	}})
	assert.NoError(t, err)
	serviceID := service.ServiceId
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID})

	t.Run("put schemas, should ok", func(t *testing.T) {
		requestBody := `{
  "schemas": [
    {
      "schemaId": "1",
      "schema": "xxx",
      "summary": "hash"
    }
  ]
}`
		r, _ := http.NewRequest(http.MethodPost, "/v4/default/registry/microservices/"+serviceID+"/schemas", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("put schemas again, should ok", func(t *testing.T) {
		requestBody := `{
  "schemas": [
    {
      "schemaId": "1",
      "schema": "yyy",
      "summary": "hash"
    }
  ]
}`
		r, _ := http.NewRequest(http.MethodPost, "/v4/default/registry/microservices/"+serviceID+"/schemas", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("put new schemas, should ok", func(t *testing.T) {
		requestBody := `{
  "schemas": [
    {
      "schemaId": "2",
      "schema": "zzz",
      "summary": "hash"
    }
  ]
}`
		r, _ := http.NewRequest(http.MethodPost, "/v4/default/registry/microservices/"+serviceID+"/schemas", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		service, err := discosvc.GetService(util.WithNoCache(ctx), &pb.GetServiceRequest{ServiceId: serviceID})
		assert.NoError(t, err)
		assert.Equal(t, []string{"2"}, service.Schemas)
	})

	t.Run("put an invalid schema, should failed", func(t *testing.T) {
		requestBody := `{
  "schemas": []
}`
		r, _ := http.NewRequest(http.MethodPost, "/v4/default/registry/microservices/"+serviceID+"/schemas", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("put schemas without summary, should failed", func(t *testing.T) {
		requestBody := `{
  "schemas": [
    {
      "schemaId": "2",
      "schema": "zzz"
    }
  ]
}`
		r, _ := http.NewRequest(http.MethodPost, "/v4/default/registry/microservices/"+serviceID+"/schemas", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSchemaRouter_ListSchema(t *testing.T) {
	ctx := context.TODO()

	service, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{Service: &pb.MicroService{
		ServiceName: "TestSchemaRouter_ListSchema",
		Schemas:     []string{"schemaID_1"},
	}})
	assert.NoError(t, err)
	serviceID := service.ServiceId
	defer discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{ServiceId: serviceID})

	t.Run("list schemas before upload content, should return only schemaID", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/v4/default/registry/microservices/"+serviceID+"/schemas", nil)
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp pb.GetAllSchemaResponse
		body, _ := ioutil.ReadAll(w.Body)
		err := json.Unmarshal(body, &resp)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.Schemas))
		assert.Equal(t, "schemaID_1", resp.Schemas[0].SchemaId)
		assert.Empty(t, resp.Schemas[0].Schema)
	})

	t.Run("list schema after put schemas, should ok", func(t *testing.T) {
		requestBody := `{
  "schemas": [
    {
      "schemaId": "schemaID_2",
      "schema": "schema_2",
      "summary": "summary2"
    }
  ]
}`
		r, _ := http.NewRequest(http.MethodPost, "/v4/default/registry/microservices/"+serviceID+"/schemas", bytes.NewBufferString(requestBody))
		w := httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		r, _ = http.NewRequest(http.MethodGet, "/v4/default/registry/microservices/"+serviceID+"/schemas", nil)
		w = httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp pb.GetAllSchemaResponse
		body, _ := ioutil.ReadAll(w.Body)
		err := json.Unmarshal(body, &resp)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.Schemas))

		schema := resp.Schemas[0]
		assert.Equal(t, "schemaID_2", schema.SchemaId)
		assert.Equal(t, "summary2", schema.Summary)
		assert.Empty(t, schema.Schema)

		r, _ = http.NewRequest(http.MethodGet, "/v4/default/registry/microservices/"+serviceID+"/schemas?withSchema=1", nil)
		w = httptest.NewRecorder()
		rest.GetRouter().ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		body, _ = ioutil.ReadAll(w.Body)
		err = json.Unmarshal(body, &resp)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.Schemas))

		schema = resp.Schemas[0]
		assert.Equal(t, "schemaID_2", schema.SchemaId)
		assert.Equal(t, "summary2", schema.Summary)
		assert.Equal(t, "schema_2", schema.Schema)
	})
}

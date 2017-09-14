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
package service

import (
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/server/infra/quota"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
)

func (s *ServiceController) GetSchemaInfo(ctx context.Context, in *pb.GetSchemaRequest) (*pb.GetSchemaResponse, error) {
	if in == nil || len(in.ServiceId) == 0 || len(in.SchemaId) == 0 {
		util.Logger().Errorf(nil, "get schema failed: invalid params.")
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request path."),
		}, nil
	}

	err := apt.Validate(in)
	if err != nil {
		util.Logger().Errorf(nil, "get schema failed, serviceId %s, schemaId %s: invalid params.", in.ServiceId, in.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}

	tenant := util.ParseTenantProject(ctx)

	opts := serviceUtil.QueryOptions(serviceUtil.WithNoCache(in.NoCache))

	if !serviceUtil.ServiceExist(ctx, tenant, in.ServiceId, opts...) {
		util.Logger().Errorf(nil, "get schema failed, serviceId %s, schemaId %s: service not exist.", in.ServiceId, in.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(tenant, in.ServiceId, in.SchemaId)
	opts = append(opts, registry.WithStrKey(key))
	resp, errDo := store.Store().Schema().Search(ctx, opts...)
	if errDo != nil {
		util.Logger().Errorf(errDo, "get schema failed, serviceId %s, schemaId %s: get schema info failed.", in.ServiceId, in.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Get schema info failed."),
		}, errDo
	}
	if resp.Count == 0 {
		util.Logger().Errorf(errDo, "get schema failed, serviceId %s, schemaId %s: schema not exists.", in.ServiceId, in.SchemaId)
		return &pb.GetSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Do not have this schema info."),
		}, nil
	}
	return &pb.GetSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get schema info successfully."),
		Schema:   util.BytesToStringWithNoCopy(resp.Kvs[0].Value),
	}, nil
}

func (s *ServiceController) DeleteSchema(ctx context.Context, request *pb.DeleteSchemaRequest) (*pb.DeleteSchemaResponse, error) {
	if request == nil || len(request.ServiceId) == 0 || len(request.SchemaId) == 0 {
		util.Logger().Errorf(nil, "delete schema failded: invalid params.")
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Invalid request path."),
		}, nil
	}
	err := apt.Validate(request)
	if err != nil {
		util.Logger().Errorf(err, "delete schema failded, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, nil
	}
	tenant := util.ParseTenantProject(ctx)

	if !serviceUtil.ServiceExist(ctx, tenant, request.ServiceId) {
		util.Logger().Errorf(nil, "delete schema failded, serviceId %s, schemaId %s: service not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Service does not exist."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(tenant, request.ServiceId, request.SchemaId)
	err, exist := serviceUtil.CheckSchemaInfoExist(ctx, key)
	if err != nil {
		util.Logger().Errorf(err, "delete schema failded, serviceId %s, schemaId %s: get schema failed.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Schema info does not exist."),
		}, err
	}
	if !exist {
		util.Logger().Errorf(nil, "delete schema failded, serviceId %s, schemaId %s: schema not exist.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Schema info does not exist."),
		}, nil
	}
	_, errDo := registry.GetRegisterCenter().Do(ctx, registry.DEL, registry.WithStrKey(key))
	if errDo != nil {
		util.Logger().Errorf(errDo, "delete schema failded, serviceId %s, schemaId %s: delete schema from etcd faild.", request.ServiceId, request.SchemaId)
		return &pb.DeleteSchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Delete schema info failed."),
		}, errDo
	}
	util.Logger().Infof("delete schema info successfully.%s", request.SchemaId)
	return &pb.DeleteSchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete schema info successfully."),
	}, nil
}

func (s *ServiceController) ModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (*pb.ModifySchemaResponse, error) {
	err, rst := s.canModifySchema(ctx, request)
	if err != nil {
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Modify schema info failed."),
		}, err
	}
	if !rst {
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Modify schema info failed, service or schemaId not exist."),
		}, nil
	}
	tenant := util.ParseTenantProject(ctx)

	ok, err := quota.QuotaPlugins[quota.QuataType]().Apply4Quotas(ctx, quota.SCHEMAQuotaType, tenant, request.ServiceId, 1)
	if err != nil {
		util.Logger().Errorf(err, "Add schema info failed, check resource num failed, %s, %s", request.ServiceId, request.SchemaId)
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Modify schema info failed, check resource num failed."),
		}, err
	}
	if !ok {
		util.Logger().Errorf(err, "Add schema info failed, reach the max size of shema, %s, %s", request.ServiceId, request.SchemaId)
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "reach the max size of shema."),
		}, nil
	}

	key := apt.GenerateServiceSchemaKey(tenant, request.ServiceId, request.SchemaId)
	_, errDo := registry.GetRegisterCenter().Do(ctx,
		registry.PUT,
		registry.WithStrKey(key),
		registry.WithStrValue(request.Schema))
	if errDo != nil {
		util.Logger().Errorf(errDo, "update schema failded, serviceId %s, schemaId %s: commit schema into etcd failed.", request.ServiceId, request.SchemaId)
		return &pb.ModifySchemaResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Modify schema info failed."),
		}, errDo
	}
	util.Logger().Infof("update schema success: serviceId %s, schemaId %s.", request.ServiceId, request.SchemaId)
	return &pb.ModifySchemaResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Modify schema info success."),
	}, nil

}

func (s *ServiceController) canModifySchema(ctx context.Context, request *pb.ModifySchemaRequest) (error, bool) {
	if request == nil || len(request.Schema) == 0 || len(request.SchemaId) == 0 || len(request.ServiceId) == 0 {
		util.Logger().Errorf(nil, "update schema failded: invalid params.")
		return nil, false
	}
	err := apt.Validate(request)
	if err != nil {
		util.Logger().Errorf(err, "update schema failded, serviceId %s, schemaId %s: invalid params.", request.ServiceId, request.SchemaId)
		return err, false
	}
	tenant := util.ParseTenantProject(ctx)
	service, err := serviceUtil.GetService(ctx, tenant, request.ServiceId)
	if err != nil {
		util.Logger().Errorf(err, "update schema failded, serviceId %s, schemaId %s: get service failed.", request.ServiceId, request.SchemaId)
		return err, false
	}
	if service == nil {
		util.Logger().Errorf(nil, "update schema failded, serviceId %s, schemaId %s: service not exist,%s", request.ServiceId, request.SchemaId)
		return nil, false
	}
	schemas := service.Schemas
	if !containsValueInSlice(schemas, request.SchemaId) {
		message := "Do not contain " + request.SchemaId + "Schema in service file."
		util.Logger().Errorf(nil, "update schema failded, serviceId %s, schemaId %s:%s", request.ServiceId, request.SchemaId, message)
		return nil, false
	}
	return nil, true
}

func containsValueInSlice(in []string, value string) bool {
	if in == nil || len(value) == 0 {
		return false
	}
	for _, i := range in {
		if i == value {
			return true
		}
	}
	return false
}

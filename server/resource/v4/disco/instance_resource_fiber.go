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

package disco

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/go-chassis/v2/pkg/codec"
	"github.com/gofiber/fiber/v2"
)

func FiberRegisterInstance(c *fiber.Ctx) error {
	request := &pb.RegisterInstanceRequest{}
	err := codec.Decode(c.Body(), request)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(c.Body())), err)
		rest.WriteFiberError(c, pb.ErrInvalidParams, "Unmarshal error")
		return nil
	}
	if request.Instance != nil {
		request.Instance.ServiceId = c.Params("serviceId")
	}

	resp, err := discosvc.RegisterInstance(c.UserContext(), request)
	if err != nil {
		log.Error("register instance failed", err)
		rest.WriteFiberServiceError(c, err)
		return nil
	}
	rest.WriteFiberResponse(c, nil, resp)
	return nil
}
func FiberSendHeartbeat(c *fiber.Ctx) error {
	request := &pb.HeartbeatRequest{
		ServiceId:  c.Params("serviceId"),
		InstanceId: c.Params("instanceId"),
	}
	err := discosvc.SendHeartbeat(c.UserContext(), request)
	if err != nil {
		rest.WriteFiberServiceError(c, err)
		return nil
	}
	rest.WriteFiberResponse(c, nil, nil)
	return nil
}
func FiberFindInstances(c *fiber.Ctx) error {
	var ids []string
	keys := c.Query("tags")
	if len(keys) > 0 {
		ids = strings.Split(keys, ",")
	}
	serviceName := c.Query("serviceName")
	request := &pb.FindInstancesRequest{
		ConsumerServiceId: c.Get("X-ConsumerId"),
		AppId:             c.Query("appId"),
		ServiceName:       serviceName,
		Alias:             serviceName,
		Environment:       c.Query("env"),
		Tags:              ids,
	}

	ctx := util.SetTargetDomainProject(c.UserContext(), c.Get("X-Domain-Name"), c.Params("project"))

	resp, err := discosvc.FindInstances(ctx, request)
	if err != nil {
		log.Error("find instances failed", err)
		rest.WriteFiberServiceError(c, err)
		return nil
	}

	iv, _ := ctx.Value(util.CtxRequestRevision).(string)
	ov, _ := ctx.Value(util.CtxResponseRevision).(string)
	c.Set(util.HeaderRev, ov)
	if len(iv) > 0 && iv == ov {
		c.Status(http.StatusNotModified)
		return nil
	}
	rest.WriteFiberResponse(c, nil, resp)
	return nil
}

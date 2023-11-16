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

package rest

import (
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/go-chassis/v2/pkg/codec"
	"github.com/gofiber/fiber/v2"
)

func WriteFiberError(c *fiber.Ctx, code int32, detail string) {
	err := discovery.NewError(code, detail)
	WriteFiberServiceError(c, err)
}
func WriteFiberServiceError(c *fiber.Ctx, err error) {
	e, ok := err.(*errsvc.Error)
	if !ok {
		WriteFiberError(c, discovery.ErrInternal, err.Error())
		return
	}
	status := e.StatusCode()
	b, err := codec.Encode(e)
	if err != nil {
		log.Error("json marshal failed", err)
		status = http.StatusInternalServerError
		b = util.StringToBytesWithNoCopy(err.Error())
	}
	c.Set(HeaderContentType, ContentTypeJSON)
	c.Status(status)
	_, err = c.Write(b)
	if err != nil {
		log.Error("write err response failed", err)
	}
}

// WriteFiberResponse writes http response
// If the resp is nil or represents success, response status is http.StatusOK,
// response content is obj.
// If the resp represents fail, response status is from the code in the
// resp, response content is from the message in the resp.
func WriteFiberResponse(c *fiber.Ctx, resp *discovery.Response, obj interface{}) {
	if resp != nil && resp.GetCode() != discovery.ResponseSuccess {
		WriteFiberError(c, resp.GetCode(), resp.GetMessage())
		return
	}

	if obj == nil {
		c.Set(HeaderContentType, ContentTypeText)
		c.Status(http.StatusOK)
		return
	}

	util.SetFiberContext(c, CtxResponseObject, obj)

	var (
		data []byte
		err  error
	)
	switch body := obj.(type) {
	case []byte:
		data = body
	default:
		data, err = codec.Encode(body)
		if err != nil {
			WriteFiberError(c, discovery.ErrInternal, err.Error())
			return
		}
	}
	c.Set(HeaderContentType, ContentTypeJSON)
	c.Status(http.StatusOK)
	_, err = c.Write(data)
	if err != nil {
		log.Error("write response failed", err)
	}
}

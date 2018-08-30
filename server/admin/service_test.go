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
package admin_test

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/admin"
	"github.com/apache/incubator-servicecomb-service-center/server/admin/model"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
)

var _ = Describe("'Admin' service", func() {
	Describe("execute 'dump' operation", func() {
		Context("when get all", func() {
			It("should be passed", func() {
				resp, err := admin.AdminServiceAPI.Dump(getContext(), &model.DumpRequest{})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(pb.Response_SUCCESS))
			})
		})
		Context("when get by domain project", func() {
			It("should be passed", func() {
				resp, err := admin.AdminServiceAPI.Dump(
					util.SetDomainProject(context.Background(), "x", "x"),
					&model.DumpRequest{})
				Expect(err).To(BeNil())
				Expect(resp.Response.Code).To(Equal(scerr.ErrForbidden))
			})
		})
	})
})

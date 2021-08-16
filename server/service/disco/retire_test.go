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
	"fmt"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/apache/servicecomb-service-center/test"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestRetireService(t *testing.T) {
	if !test.IsETCD() {
		return
	}

	const serviceIDPrefix = "TestRetireMicroservice"
	ctx := getContext()

	t.Run("normal case should return ok", func(t *testing.T) {
		const count = 5
		for i := 0; i < count; i++ {
			idx := fmt.Sprintf("%d", i)
			_, err := discosvc.RegisterService(ctx, &pb.CreateServiceRequest{
				Service: &pb.MicroService{
					ServiceId:   serviceIDPrefix + idx,
					ServiceName: serviceIDPrefix,
					Version:     "1.0." + idx,
				},
			})
			assert.NoError(t, err)
		}
		defer func() {
			for i := 0; i < count; i++ {
				idx := fmt.Sprintf("%d", i)
				discosvc.UnregisterService(ctx, &pb.DeleteServiceRequest{
					ServiceId: serviceIDPrefix + idx,
				})
			}
		}()

		err := discosvc.RetireService(ctx, &datasource.RetirePlan{Interval: 0, Reserve: 1})
		assert.NoError(t, err)

		resp, err := datasource.GetMetadataManager().ListServiceDetail(ctx, &pb.GetServicesInfoRequest{
			ServiceName: serviceIDPrefix,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(resp.AllServicesDetail))
		assert.Equal(t, serviceIDPrefix+"4", resp.AllServicesDetail[0].MicroService.ServiceId)
	})
}

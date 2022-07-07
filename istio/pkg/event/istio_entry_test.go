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

package event

import (
	"fmt"
	"testing"

	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"

	istioAPI "istio.io/api/networking/v1alpha3"
	istioClient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertIstioEntryToMicroservice(t *testing.T) {
	in := mockNewIstioEntry()
	out := in.Convert()

	t.Run("a serviceentry should be convertable", func(t *testing.T) {
		assert.NotNil(t, out)
	})

	me := out.(*MicroserviceEntry)

	t.Run("the prefix of istio entry service name should be automatically removed", func(t *testing.T) {
		assert.Equal(t, "test-svc", me.MicroService.ServiceName)
	})

	t.Run("the number of microservice instances should be the same as the number of serviceentry hosts", func(t *testing.T) {
		assert.Equal(t, len(in.ServiceEntry.Spec.Hosts), len(me.Instances))
	})

	t.Run("converted microservice instance should be as same as istio workload entry", func(t *testing.T) {
		for i, host := range in.ServiceEntry.Spec.Hosts {
			endpoints := []string{
				fmt.Sprintf("rest://%s:%d", host, 1234),
			}
			expectInstance := &InstanceEntry{
				&discovery.MicroServiceInstance{
					HostName:  host,
					Endpoints: endpoints,
				},
			}

			assert.Equal(t, expectInstance, me.Instances[i])
		}
	})
}

func mockNewIstioEntry() ServiceEntry {
	se := ServiceEntry{
		ServiceEntry: &istioClient.ServiceEntry{
			ObjectMeta: k8s.ObjectMeta{
				Name: "synthetic-test-svc",
			},
			Spec: istioAPI.ServiceEntry{
				Hosts: []string{"api.test-svc-1.com", "api.test-svc-2.com"},
				Ports: []*istioAPI.Port{
					{
						Name:       "http",
						Number:     80,
						Protocol:   "HTTP",
						TargetPort: 1234,
					},
				},
			},
		},
	}
	return se
}

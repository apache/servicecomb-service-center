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

package event_test

import (
	"fmt"
	"testing"

	"github.com/apache/servicecomb-service-center/istio/pkg/event"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"

	istioAPI "istio.io/api/networking/v1alpha3"
	istioClient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mockNewIstioServiceEntryNoWorkloads() event.ServiceEntry {
	se := event.ServiceEntry{
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

func TestConvertServiceEntryToMicroserviceNoWorkloads(t *testing.T) {
	in := mockNewIstioServiceEntryNoWorkloads()
	inIstio := in.ServiceEntry
	out := in.Convert().(*event.MicroserviceEntry)

	expectName := "test-svc"
	assert.Equal(t, expectName, out.MicroService.ServiceName)

	assert.Equal(t, len(inIstio.Spec.Hosts), len(out.Instances))

	for i, host := range inIstio.Spec.Hosts {
		endpoints := []string{
			fmt.Sprintf("rest://%s:%d", host, 1234),
		}
		expectInstance := &event.InstanceEntry{
			&discovery.MicroServiceInstance{
				HostName:  host,
				Endpoints: endpoints,
			},
		}

		assert.Equal(t, expectInstance, out.Instances[i])
	}
}

func TestNewServiceEntry(t *testing.T) {
	inCurr := &istioClient.ServiceEntry{}
	out := event.NewServiceEntry(inCurr)
	expectServiceEntry := &event.ServiceEntry{
		ServiceEntry:    inCurr,
		WorkloadEntries: []*event.WorkloadEntry{},
	}
	assert.Equal(t, expectServiceEntry, out)
}

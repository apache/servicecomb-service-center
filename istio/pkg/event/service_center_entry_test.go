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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-chassis/cari/discovery"
	istioAPI "istio.io/api/networking/v1alpha3"
)

func TestConvertMicroserviceToIstioMultipleInstance(t *testing.T) {
	in := mockNewMicroserviceEntryMultipleInstance()
	out := in.Convert()
	t.Run("a microservice with multiple instances should be convertable", func(t *testing.T) {
		assert.NotNil(t, out)
	})

	sn := out.(*ServiceEntry)
	t.Run("the microservice name should be converted to lowercase in istio service entry", func(t *testing.T) {
		assert.Equal(t, strings.ToLower(in.MicroService.ServiceName), sn.ServiceEntry.ObjectMeta.Name)
	})

	t.Run("the microservice name should be converted to lowercase in istio service entry", func(t *testing.T) {
		assert.Equal(t, 1, len(sn.ServiceEntry.Spec.Hosts))
		assert.Equal(t, strings.ToLower(in.MicroService.ServiceName), sn.ServiceEntry.Spec.Hosts[0])
	})

	t.Run("the number of microservice instances should be the same as the number of converted serviceentry endpoints", func(t *testing.T) {
		assert.Equal(t, len(in.Instances), len(sn.ServiceEntry.Spec.Endpoints))
	})

	t.Run("converted serviceentry port should be as same as that in microservice instance", func(t *testing.T) {
		expectPort1 := &istioAPI.Port{Number: 1111, Protocol: "HTTP", Name: "HTTP-1111", TargetPort: 1111}
		expectPort2 := &istioAPI.Port{Number: 2222, Protocol: "HTTPS", Name: "HTTPS-2222", TargetPort: 2222}
		expectPort3 := &istioAPI.Port{Number: 3333, Protocol: "HTTP", Name: "HTTP-3333", TargetPort: 3333}
		expectPorts := []*istioAPI.Port{expectPort1, expectPort2, expectPort3}
		expectPortsMap := map[string]*istioAPI.Port{expectPort1.Name: expectPort1, expectPort2.Name: expectPort2, expectPort3.Name: expectPort3}
		assert.Equal(t, len(expectPorts), len(sn.ServiceEntry.Spec.Ports))
		for _, outPort := range sn.ServiceEntry.Spec.Ports {
			assert.Contains(t, expectPortsMap, outPort.Name)
			assert.Equal(t, expectPortsMap[outPort.Name], outPort)
		}
	})

	t.Run("the converted workload entry should be exactly same as microservice instance", func(t *testing.T) {
		for i, inst := range in.Instances {
			d := i + 1
			expectWorkloadEntry := istioAPI.WorkloadEntry{
				Address: fmt.Sprintf(`%d.%d.%d.%d`, d, d, d, d),
				Ports: map[string]uint32{
					"HTTP-1111":  1111,
					"HTTPS-2222": 2222,
					"HTTP-3333":  3333,
				},
				Labels: map[string]string{
					"instanceId": inst.InstanceId,
					"name":       inst.HostName,
				},
			}
			assert.Equal(t, expectWorkloadEntry, *sn.ServiceEntry.Spec.Endpoints[i])
		}
	})
}

func TestConvertMicroserviceToIstioInvalidEndpoints(t *testing.T) {
	in := mockNewMicroserviceEntryInvalidEndpoints()
	out := in.Convert()
	t.Run("a microservice with invalid endpoints should be convertable", func(t *testing.T) {
		assert.NotNil(t, out)
	})

	sn := out.(*ServiceEntry)

	t.Run("the microservice name should be converted to lowercase in istio service entry", func(t *testing.T) {
		assert.Equal(t, strings.ToLower(in.MicroService.ServiceName), sn.ServiceEntry.ObjectMeta.Name)
	})

	t.Run("the microservice name should be converted to lowercase in istio service entry", func(t *testing.T) {
		assert.Equal(t, 1, len(sn.ServiceEntry.Spec.Hosts))
		assert.Equal(t, strings.ToLower(in.MicroService.ServiceName), sn.ServiceEntry.Spec.Hosts[0])
	})

	t.Run("given 3 instance with 2 invalidate, should only convert the one that validate", func(t *testing.T) {
		assert.Equal(t, 1, len(sn.ServiceEntry.Spec.Endpoints))
	})

	t.Run("converted serviceentry port should be as same as that in microservice instance", func(t *testing.T) {
		expectPort1 := &istioAPI.Port{Number: 1111, Protocol: "HTTP", Name: "HTTP-1111", TargetPort: 1111}
		expectPort2 := &istioAPI.Port{Number: 1111, Protocol: "HTTPS", Name: "HTTPS-1111", TargetPort: 1111}
		expectPort3 := &istioAPI.Port{Number: 2222, Protocol: "HTTP", Name: "HTTP-2222", TargetPort: 2222}
		expectPort4 := &istioAPI.Port{Number: 3333, Protocol: "GRPC", Name: "GRPC-3333", TargetPort: 3333}
		expectPorts := []*istioAPI.Port{expectPort1, expectPort2, expectPort3, expectPort4}
		expectPortsMap := map[string]*istioAPI.Port{expectPort1.Name: expectPort1, expectPort2.Name: expectPort2, expectPort3.Name: expectPort3, expectPort4.Name: expectPort4}
		assert.Equal(t, len(expectPorts), len(sn.ServiceEntry.Spec.Ports))
		for _, outPort := range sn.ServiceEntry.Spec.Ports {
			assert.Contains(t, expectPortsMap, outPort.Name)
			assert.Equal(t, expectPortsMap[outPort.Name], outPort)
		}
	})

	t.Run("the converted workload entry should be exactly same as the validate microservice instance", func(t *testing.T) {
		d := 1
		expectWorkloadEntry := istioAPI.WorkloadEntry{
			Address: fmt.Sprintf(`%d.%d.%d.%d`, d, d, d, d),
			Ports: map[string]uint32{
				"HTTP-1111":  1111,
				"HTTPS-1111": 1111,
				"HTTP-2222":  2222,
				"GRPC-3333":  3333,
			},
			Labels: map[string]string{
				"instanceId": in.Instances[0].InstanceId,
				"name":       in.Instances[0].HostName,
			},
		}
		assert.Equal(t, expectWorkloadEntry, *sn.ServiceEntry.Spec.Endpoints[0])
	})
}

func TestConvertMicroserviceToIstioNoInstances(t *testing.T) {
	in := mockNewMicroserviceEntryNoInstances()
	out := in.Convert()
	t.Run("a microservice with no instance should be convertable", func(t *testing.T) {
		assert.NotNil(t, out)
	})

	sn := out.(*ServiceEntry)

	t.Run("the microservice name should be converted to lowercase in istio service entry", func(t *testing.T) {
		assert.Equal(t, strings.ToLower(in.MicroService.ServiceName), sn.ServiceEntry.ObjectMeta.Name)
	})

	t.Run("the microservice name should be converted to lowercase in istio service entry", func(t *testing.T) {
		assert.Equal(t, 1, len(sn.ServiceEntry.Spec.Hosts))
		assert.Equal(t, strings.ToLower(in.MicroService.ServiceName), sn.ServiceEntry.Spec.Hosts[0])
	})

	t.Run("given no instances, the converted serviceEntry should have 0 endpoints and ports", func(t *testing.T) {
		assert.Equal(t, 0, len(sn.ServiceEntry.Spec.Endpoints))
		assert.Equal(t, 0, len(sn.ServiceEntry.Spec.Ports))
	})
}

func mockNewMicroserviceEntryNoInstances() MicroserviceEntry {
	ms := &discovery.MicroService{
		ServiceId:   "testsvc123456",
		AppId:       "Test-App",
		ServiceName: "Test-Svc",
	}
	return MicroserviceEntry{
		MicroService: ms,
	}
}

func mockNewMicroserviceEntryInvalidEndpoints() MicroserviceEntry {
	ms := &discovery.MicroService{
		ServiceId:   "testsvc123456",
		AppId:       "Test-App",
		ServiceName: "Test-Svc",
	}
	var insts []*InstanceEntry
	ie := &InstanceEntry{
		&discovery.MicroServiceInstance{
			ServiceId:  "testsvc123456",
			InstanceId: "testinst1a",
			Endpoints: []string{
				"rest://1.1.1.1:1111",
				"rest://1.1.1.1:1111?sslEnabled=true",
				"foo://2.2.2.2:2222",  // Will cause warning; unsupported protocol, will default to http
				"grpc://1.1.1.1:3333", // GRPC is also supported by istio
			},
			HostName: "test-svc-host-1",
		},
	}
	ie1 := &InstanceEntry{
		&discovery.MicroServiceInstance{
			// Will fail to convert
			ServiceId:  "testsvc123456",
			InstanceId: "testinst3c",
			Endpoints: []string{
				"@@@@://4.4.4.4:1111", // Will cause instance to be skipped; invalid uri (protocol)
			},
			HostName: "test-svc-host-3",
		},
	}
	ie2 := &InstanceEntry{
		&discovery.MicroServiceInstance{
			// Will fail to convert
			ServiceId:  "testsvc123456",
			InstanceId: "testinst4d",
			Endpoints:  []string{}, // Will cause instance to be skipped; no endpoints
			HostName:   "test-svc-host-4",
		},
	}
	insts = append(insts, ie, ie1, ie2)

	return MicroserviceEntry{
		MicroService: ms,
		Instances:    insts,
	}
}

func mockNewMicroserviceEntryMultipleInstance() MicroserviceEntry {
	ms := &discovery.MicroService{
		ServiceId:   "testsvc123456",
		AppId:       "Test-App",
		ServiceName: "Test-Svc",
	}

	var insts []*InstanceEntry
	ie := &InstanceEntry{
		&discovery.MicroServiceInstance{
			ServiceId:  "testsvc123456",
			InstanceId: "testinst1a",
			Endpoints:  []string{"rest://1.1.1.1:1111", "rest://1.1.1.1:2222?sslEnabled=true", "rest://1.1.1.1:3333"},
			HostName:   "test-svc-host-1",
		},
	}
	ie1 := &InstanceEntry{
		&discovery.MicroServiceInstance{
			ServiceId:  "testsvc123456",
			InstanceId: "testinst2b",
			Endpoints:  []string{"rest://2.2.2.2:1111", "rest://2.2.2.2:2222?sslEnabled=true", "rest://2.2.2.2:3333"},
			HostName:   "test-svc-host-2",
		},
	}
	insts = append(insts, ie, ie1)

	return MicroserviceEntry{
		MicroService: ms,
		Instances:    insts,
	}
}

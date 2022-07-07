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
	"net/url"
	"strings"

	"github.com/go-chassis/cari/discovery"
	istioAPI "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
)

// An Istio WorkloadEntry and the known ports of its parent ServiceEntry.
type WorkloadEntry struct {
	*istioAPI.WorkloadEntry
	// A subset of ports used by the WorkloadEntry's parent ServiceEntry.
	// Used in conjunction with other WorkloadEntry's ServicePorts to construct a complete set of the ServiceEntry's ports.
	ServicePorts []*istioAPI.Port
}

// Convert an Istio WorkloadEntry to a service center Microservice Instance event.
func (c *WorkloadEntry) Convert() Event {
	inst := convertIstioAddressToMicroserviceInstance(c.Address, c.ServicePorts)
	if id, ok := c.Labels[InstanceIdLabel]; ok {
		// WorkloadEntry was previously converted from service center Microservice instance; restore its id
		inst.InstanceId = id
	}
	return inst
}

// Convert an Istio host/endpoint's address and ports to a service center Microservice Instance entry.
func convertIstioAddressToMicroserviceInstance(address string, ports []*istioAPI.Port) *InstanceEntry {
	endpoints := []string{}
	// Create equivalent Microservice endpoints for Istio address+port combinations
	for _, port := range ports {
		ep := getMicroserviceEndpointFromIstio(address, port)
		endpoints = append(endpoints, ep)
	}
	return &InstanceEntry{
		MicroServiceInstance: &discovery.MicroServiceInstance{
			Endpoints: endpoints,
			// we.Address could be a hostname if using DNS
			HostName: address,
		},
	}
}

// An Istio ServiceEntry and its associated WorkloadEntry(s).
type ServiceEntry struct {
	// An Istio ServiceEntry struct supported by the Istio client library.
	ServiceEntry *v1alpha3.ServiceEntry
	// WorkloadEntry(s) representing instances of the Istio ServiceEntry.
	WorkloadEntries []*WorkloadEntry
}

func NewServiceEntry(curr *v1alpha3.ServiceEntry) *ServiceEntry {
	wles := []*WorkloadEntry{}
	for _, wle := range curr.Spec.Endpoints {
		// Record ServiceEntry's ports in each WorkloadEntry
		wles = append(wles, &WorkloadEntry{WorkloadEntry: wle, ServicePorts: curr.Spec.Ports})
	}
	return &ServiceEntry{
		ServiceEntry:    curr,
		WorkloadEntries: wles,
	}
}

// Convert an Istio ServiceEntry to a MicroService event
func (s *ServiceEntry) Convert() Event {
	name := getMicroserviceNameFromIstio(s.ServiceEntry.Name)
	ms := &discovery.MicroService{
		ServiceName: name,
	}
	insts := []*InstanceEntry{}

	if len(s.ServiceEntry.Spec.Endpoints) > 0 {
		// Convert a ServiceEntry with WorkloadEntry(s)
		for _, wle := range s.WorkloadEntries {
			inst := wle.Convert()
			insts = append(insts, inst.(*InstanceEntry))
		}
	} else if len(s.ServiceEntry.Spec.Ports) > 0 {
		// Convert a ServiceEntry with Host(s) and Port(s) only
		for _, host := range s.ServiceEntry.Spec.Hosts {
			ports := s.ServiceEntry.Spec.GetPorts()
			// Construct a Service center instance using only a host address and its ports
			inst := convertIstioAddressToMicroserviceInstance(host, ports)
			insts = append(insts, inst)
		}
	}
	return &MicroserviceEntry{
		MicroService: ms,
		Instances:    insts,
	}
}

// Reports whether a protocol string is supported by Service center
func isMicroserviceRestProtocol(protocol string) bool {
	return strings.Contains("TCP HTTP HTTPS", protocol)
}

// Construct an endpoint for a Service center MicroServiceInstance from an Istio address and port
func getMicroserviceEndpointFromIstio(address string, port *istioAPI.Port) string {
	u := url.URL{}
	if isMicroserviceRestProtocol(port.Protocol) {
		if port.Protocol == "HTTPS" {
			// Enable SSL for the Service center REST endpoint
			u.RawQuery = EnableSSL
		}
		u.Scheme = "rest"
	}
	// ServiceEntry's "internal" port used by its corresponding WorkloadEntry
	var targetPort uint32
	if port.TargetPort != 0 {
		targetPort = port.TargetPort
	} else {
		// Use "external" ServiceEntry port instead
		targetPort = port.Number
	}
	// Service center endpoint string uses format <protocol>://<host>:<port>
	u.Host = fmt.Sprintf("%s:%d", address, targetPort)
	return u.String()
}

// Construct a Service center service name from an Istio service name
func getMicroserviceNameFromIstio(svcName string) string {
	svcName = strings.TrimPrefix(svcName, "synthetic-")
	// No need to return titlecased service names
	// (example: https://github.com/apache/servicecomb-service-center/blob/6f26aaa7698691d40e17c6644ac71d51b6770772/integration/microservices_test.go#L592)
	return svcName
}

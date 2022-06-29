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
	"strconv"
	"strings"

	"github.com/go-chassis/cari/discovery"
	istioAPI "istio.io/api/networking/v1alpha3"
	istioClient "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/pkg/config/protocol"
	"istio.io/pkg/log"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// microservice protocol used for REST endpoints, equivalent to HTTP/HTTPS.
	RestProtocol = "rest"
	// Istio label corresponding to id of original microservice instance before being converted to WorkloadEntry.
	InstanceIdLabel = "instanceId"
	// microservice endpoint query string that enables SSL.
	EnableSSL = "sslEnabled=true"
)

type ChangeEvent struct {
	// Type of change event
	Action discovery.EventType
	// Event payload
	Event
}

type InstanceEntry struct {
	*discovery.MicroServiceInstance
}

// A Service center MicroService and its associated instance(s).
type MicroserviceEntry struct {
	// Microservice struct
	MicroService *discovery.MicroService
	// Instances of the MicroService
	Instances []*InstanceEntry
}

// Convert a MicroServiceInstance to an Istio WorkloadEntry event.
func (c *InstanceEntry) Convert() Event {
	// Istio ServiceEntry port names mapped to the "internal" port number of their target WorkloadEntry
	portMap := map[string]uint32{}
	// Istio ServiceEntry port names mapped to port structs
	svcPortMap := map[string]*istioAPI.Port{}
	var address string
	if len(c.Endpoints) == 0 {
		log.Errorf("Service center Microservice Instance %v has no endpoint found\n", c.InstanceId)
		return nil
	}
	for i, ep := range c.Endpoints {
		u, err := url.Parse(ep)
		if err != nil {
			log.Errorf("Service center Microservice Instance %v has endpoint \"%v\" that could not be parsed as a valid URL\n", c.InstanceId, ep)
			return nil
		}
		if i == 0 {
			// Only using first endpoint's hostname as the WorkloadEntry's address, Service center Microservice expects all endpoints for an instance to share a hostname
			address = u.Hostname()
		}
		istioProtocol := getIstioProtocolFromURL(*u)
		port, _ := strconv.Atoi(u.Port())
		// Name for Istio ports are set to <protocol>-<portNumber>
		portName := fmt.Sprintf("%s-%s", istioProtocol, u.Port())
		svcPortMap[portName] = newIstioPort(portName, istioProtocol, uint32(port))
		portMap[portName] = uint32(port)
	}
	// Convert ServiceEntry port map to array
	svcPorts := []*istioAPI.Port{}
	for _, port := range svcPortMap {
		svcPorts = append(svcPorts, port)
	}
	return &WorkloadEntry{
		WorkloadEntry: &istioAPI.WorkloadEntry{
			Address: address,
			Ports:   portMap,
			Labels: map[string]string{
				"name":          c.HostName,
				InstanceIdLabel: c.InstanceId,
			},
		},
		ServicePorts: svcPorts,
	}
}

// Convert a MicroService to an Istio ServiceEntry.
func (c *MicroserviceEntry) Convert() Event {
	ms := c.MicroService
	insts := c.Instances
	name := strings.ToLower(ms.ServiceName)

	location := istioAPI.ServiceEntry_MESH_INTERNAL
	resolution := istioAPI.ServiceEntry_STATIC
	wles := []*istioAPI.WorkloadEntry{}
	// Ensures that only one port struct exists for each port name
	svcPortMap := map[string]*istioAPI.Port{}
	for _, inst := range insts {
		wle := inst.Convert()
		if wle != nil {
			workloadEntry := wle.(*WorkloadEntry)
			for _, istioPort := range workloadEntry.ServicePorts {
				svcPortMap[istioPort.Name] = istioPort
			}
			wles = append(wles, workloadEntry.WorkloadEntry)
		}
	}

	// Convert port map to array representing all ports exposed by ServiceEntry
	svcPorts := []*istioAPI.Port{}
	for _, port := range svcPortMap {
		svcPorts = append(svcPorts, port)
	}
	se := &istioClient.ServiceEntry{
		ObjectMeta: k8s.ObjectMeta{
			Name: name,
		},
		Spec: istioAPI.ServiceEntry{
			Hosts:      []string{name},
			Ports:      svcPorts,
			Location:   location,
			Resolution: resolution,
			Endpoints:  wles,
		},
	}
	return NewServiceEntry(se)
}

func getIstioProtocolFromURL(u url.URL) string {
	p := u.Scheme
	if p == RestProtocol {
		// Microservice uses query string to signify HTTPS endpoint
		if strings.Contains(u.RawQuery, EnableSSL) {
			p = "https"
		} else {
			p = "http"
		}
	}
	istioP := protocol.Parse(p)
	if istioP.IsUnsupported() {
		istioP = protocol.HTTP
	}
	return string(istioP)
}

// Construct an Istio ServiceEntry port.
func newIstioPort(name string, protocol string, number uint32) *istioAPI.Port {
	return &istioAPI.Port{
		Number:     number,
		TargetPort: number,
		Protocol:   protocol,
		Name:       name,
	}
}

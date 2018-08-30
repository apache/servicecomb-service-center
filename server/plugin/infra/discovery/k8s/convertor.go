// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	"k8s.io/api/core/v1"
	"net/url"
	"strconv"
	"strings"
)

func getLabel(labels map[string]string, key, def string) string {
	if v, ok := labels[key]; ok {
		return v
	}
	return def
}

func getRegionAZ(node *v1.Node) (string, string) {
	region, exists := node.Labels[LabelNodeRegion]
	if !exists {
		return "", ""
	}
	zone, exists := node.Labels[LabelNodeAZ]
	if !exists {
		return region, ""
	}
	return region, zone
}

func getFullName(namespace, name string) string {
	if len(namespace) != 0 {
		return namespace + "/" + name
	}
	return name
}

func getProtocol(port v1.EndpointPort) string {
	switch port.Protocol {
	case SchemaTCP:
		return SchemaHTTP
	default:
		return strings.ToLower(string(port.Protocol))
	}
}

func generateEndpoint(ip string, port v1.EndpointPort) string {
	u := url.URL{
		Scheme: getProtocol(port),
		Host:   ip + ":" + strconv.FormatInt(int64(port.Port), 10),
	}
	return u.String()
}

func generateServiceKey(domainProject string, svc *v1.Service) *pb.MicroServiceKey {
	return &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: getLabel(svc.Labels, LabelEnvironment, ""),
		AppId:       getLabel(svc.Labels, LabelApp, pb.APP_ID),
		ServiceName: svc.Name,
		Version:     getLabel(svc.Labels, LabelVersion, pb.VERSION),
	}
}

func FromK8sService(svc *v1.Service) (ms *pb.MicroService) {
	ms = &pb.MicroService{
		ServiceId:   string(svc.UID),
		Environment: getLabel(svc.Labels, LabelEnvironment, ""),
		AppId:       getLabel(svc.Labels, LabelApp, pb.APP_ID),
		ServiceName: svc.Name,
		Version:     getLabel(svc.Labels, LabelVersion, pb.VERSION),
		Level:       "BACK",
		Status:      pb.MS_UP,
		Framework: &pb.FrameWorkProperty{
			Name: Name,
		},
		RegisterBy: pb.REGISTERBY_PLATFORM,
		Properties: map[string]string{
			PropNamespace:    svc.Namespace,
			PropServiceType:  string(svc.Spec.Type),
			PropExternalName: svc.Spec.ExternalName,
		},
	}
	ms.Timestamp = strconv.FormatInt(svc.CreationTimestamp.Unix(), 10)
	ms.ModTimestamp = ms.Timestamp
	return
}

func AsKeyValue(key string, v interface{}, resourceVersion string) *discovery.KeyValue {
	rev, _ := strconv.ParseInt(resourceVersion, 10, 64)
	return &discovery.KeyValue{
		Key:         util.StringToBytesWithNoCopy(key),
		Value:       v,
		ModRevision: rev,
	}
}

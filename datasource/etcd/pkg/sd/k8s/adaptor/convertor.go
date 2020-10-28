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

package adaptor

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/pkg/sd"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
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

func getProtocol(port v1.EndpointPort) (string, bool) {
	switch port.Protocol {
	case SchemaTCP:
		switch strings.ToUpper(port.Name) {
		case SchemaHTTPS:
			return protocolRest, true
		default:
			return protocolRest, false
		}
	default:
		return strings.ToLower(string(port.Protocol)), false
	}
}

func generateEndpoint(ip string, port v1.EndpointPort) string {
	protocol, secure := getProtocol(port)
	u := url.URL{
		Scheme: protocol,
		Host:   ip + ":" + strconv.FormatInt(int64(port.Port), 10),
	}
	if secure {
		u.RawQuery = "sslEnabled=true"
	}
	return u.String()
}

func generateServiceKey(domainProject string, svc *v1.Service) *pb.MicroServiceKey {
	return &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: getLabel(svc.Labels, LabelEnvironment, ""),
		AppId:       getLabel(svc.Labels, LabelApp, pb.AppID),
		ServiceName: svc.Name,
		Version:     getLabel(svc.Labels, LabelVersion, pb.VERSION),
	}
}

func FromK8sService(domainProject string, svc *v1.Service) (ms *pb.MicroService) {
	ms = &pb.MicroService{
		ServiceId:   generateServiceID(domainProject, svc),
		Environment: getLabel(svc.Labels, LabelEnvironment, ""),
		AppId:       getLabel(svc.Labels, LabelApp, pb.AppID),
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

func AsKeyValue(key string, v interface{}, resourceVersion string) *sd.KeyValue {
	rev, _ := strconv.ParseInt(resourceVersion, 10, 64)
	kv := sd.NewKeyValue()
	kv.Key = util.StringToBytesWithNoCopy(key)
	kv.Value = v
	kv.ModRevision = rev
	kv.ClusterName = etcd.Configuration().ClusterName
	return kv
}

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
	"context"
	"github.com/apache/servicecomb-service-center/pkg/queue"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/uuid"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
	"time"
)

const (
	Name                 = "Kubernetes"
	TypeService  K8sType = "Service"
	TypeEndpoint K8sType = "Endpoints"
	TypeNode     K8sType = "Node"
	TypePod      K8sType = "Pod"

	SchemaTCP    = "TCP"
	SchemaHTTP   = "HTTP"
	SchemaHTTPS  = "HTTPS"
	protocolRest = "rest"

	defaultDomainProject = "default/default"

	// k8s labels
	LabelApp         = "app"
	LabelVersion     = "version"
	LabelEnvironment = "environment"
	LabelNodeRegion  = "failure-domain.beta.kubernetes.io/region"
	LabelNodeAZ      = "failure-domain.beta.kubernetes.io/zone"

	// register annotations
	AnnotationRegister = "service-center.servicecomb.io/register"

	// properties
	PropNamespace    = "namespace"
	PropExternalName = "externalName"
	PropServiceType  = "type"
	PropNodeIP       = "nodeIP"

	minWaitInterval       = 5 * time.Second
	defaultResyncInterval = 60 * time.Second
	eventQueueSize        = 1000
)

var (
	closedCh    = make(chan struct{})
	eventQueues = util.NewConcurrentMap(0)
)

func init() {
	close(closedCh)
}

func createKubeClient(kubeConfigPath string) (kubernetes.Interface, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func Queue(t K8sType) *queue.TaskQueue {
	q, _ := eventQueues.Fetch(t, func() (interface{}, error) {
		q := queue.NewTaskQueue(eventQueueSize)
		q.Run()
		return q, nil
	})
	return q.(*queue.TaskQueue)
}

func ShouldRegisterService(service *v1.Service) bool {
	if service.Namespace == meta.NamespaceSystem {
		return false
	}
	if register, ok := service.ObjectMeta.Annotations[AnnotationRegister]; ok && register == "false" {
		return false
	}
	return true
}

func UUID(id types.UID) string {
	return strings.Replace(string(id), "-", "", -1)
}

func generateServiceId(domainProject string, svc *v1.Service) string {
	indexKey := core.GenerateServiceIndexKey(generateServiceKey(domainProject, svc))
	ctx := context.WithValue(context.Background(), uuid.ContextKey, indexKey)
	return mgr.Plugins().UUID().GetServiceId(ctx)
}

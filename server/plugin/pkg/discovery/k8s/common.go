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
	"github.com/apache/incubator-servicecomb-service-center/pkg/queue"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

	// properties
	PropNamespace    = "namespace"
	PropExternalName = "externalName"
	PropServiceType  = "type"

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

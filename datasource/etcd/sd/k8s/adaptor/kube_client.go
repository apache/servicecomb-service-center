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
	"fmt"
	"os"
	"sync"
	"time"

	pb "github.com/go-chassis/cari/discovery"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm"
)

var (
	client     *K8sClient
	clientOnce sync.Once
)

type K8sType string

type K8sClient struct {
	// eventFuncs is used for store functions called k8s event handler
	eventFuncs util.ConcurrentMap
	// ipIndex is used for index pod object; key is ip, value is pod full name
	ipIndex util.ConcurrentMap

	kubeClient kubernetes.Interface
	services   ListWatcher
	endpoints  ListWatcher
	nodes      ListWatcher
	pods       ListWatcher

	ready     chan struct{}
	stopCh    chan struct{}
	goroutine *gopool.Pool
}

func (c *K8sClient) init() (err error) {
	c.ready = make(chan struct{})
	c.stopCh = make(chan struct{})
	c.goroutine = gopool.New(context.Background())

	// if KUBERNETES_CONFIG_PATH is unset, then service center must be deployed in the same k8s cluster
	c.kubeClient, err = createKubeClient(os.Getenv("KUBERNETES_CONFIG_PATH"))
	if err != nil {
		log.Error("create kube client failed", err)
		return
	}

	// if KUBERNETES_NAMESPACE is unset, then list watch all namespaces
	listerFactory := informers.NewFilteredSharedInformerFactory(
		c.kubeClient, defaultResyncInterval, os.Getenv("KUBERNETES_NAMESPACE"), nil)
	c.services = c.newListWatcher(TypeService, listerFactory.Core().V1().Services().Informer())
	c.endpoints = c.newListWatcher(TypeEndpoint, listerFactory.Core().V1().Endpoints().Informer())
	c.nodes = c.newListWatcher(TypeNode, listerFactory.Core().V1().Nodes().Informer())
	c.pods = c.newListWatcher(TypePod, listerFactory.Core().V1().Pods().Informer())

	// append ipIndex build function
	c.AppendEventFunc(TypePod, c.onPodEvent)
	return
}

func (c *K8sClient) newListWatcher(t K8sType, lister cache.SharedIndexInformer) (lw ListWatcher) {
	lw = NewListWatcher(t, lister, c.getEvent(t))
	return
}

func (c *K8sClient) getEvent(t K8sType) OnEventFunc {
	return func(evt K8sEvent) {
		fs, ok := c.eventFuncs.Get(t)
		if !ok {
			return
		}
		for _, f := range fs.([]OnEventFunc) {
			f(evt)
		}
	}
}

// onPodEvent is method to build ipIndex
func (c *K8sClient) onPodEvent(evt K8sEvent) {
	pod, ok := evt.Object.(*v1.Pod)
	if !ok {
		deletedState, ok := evt.Object.(cache.DeletedFinalStateUnknown)
		if !ok {
			log.Warn(fmt.Sprintf("event object is not a pod %#v", evt.Object))
			return
		}
		pod, ok = deletedState.Obj.(*v1.Pod)
		if !ok {
			log.Warn(fmt.Sprintf("deletedState is not a pod %#v", evt.Object))
			return
		}
	}

	if len(pod.Status.PodIP) == 0 {
		return
	}

	podName := getFullName(pod.Namespace, pod.Name)
	switch evt.EventType {
	case pb.EVT_CREATE, pb.EVT_UPDATE:
		switch pod.Status.Phase {
		case v1.PodPending, v1.PodRunning:
			c.ipIndex.Put(pod.Status.PodIP, podName)
		default:
		}
	case pb.EVT_DELETE:
		c.ipIndex.Remove(pod.Status.PodIP)
	}
}

// unsafe
func (c *K8sClient) AppendEventFunc(t K8sType, f OnEventFunc) {
	itf, _ := c.eventFuncs.Fetch(t, func() (interface{}, error) {
		return []OnEventFunc{}, nil
	})
	fs := itf.([]OnEventFunc)
	fs = append(fs, f)
	c.eventFuncs.Put(t, fs)
}

func (c *K8sClient) waitForSync(lw ListWatcher) ListWatcher {
	<-c.ready
	return lw
}

func (c *K8sClient) Services() ListWatcher {
	return c.waitForSync(c.services)
}

func (c *K8sClient) Endpoints() ListWatcher {
	return c.waitForSync(c.endpoints)
}

func (c *K8sClient) Pods() ListWatcher {
	return c.waitForSync(c.pods)
}

func (c *K8sClient) Nodes() ListWatcher {
	return c.waitForSync(c.nodes)
}

func (c *K8sClient) GetDomainProject() string {
	return defaultDomainProject
}

func (c *K8sClient) GetService(namespace, name string) (svc *v1.Service) {
	obj, ok, err := c.Services().GetStore().GetByKey(getFullName(namespace, name))
	if err != nil {
		log.Error(fmt.Sprintf("get k8s service[%s/%s] failed", namespace, name), err)
		return
	}
	if !ok {
		return
	}
	svc = obj.(*v1.Service)
	return
}

func (c *K8sClient) GetEndpoints(namespace, name string) (ep *v1.Endpoints) {
	obj, ok, err := c.Endpoints().GetStore().GetByKey(getFullName(namespace, name))
	if err != nil {
		log.Error(fmt.Sprintf("get k8s endpoints[%s/%s] failed", namespace, name), err)
		return
	}
	if !ok {
		return
	}
	ep = obj.(*v1.Endpoints)
	return
}

func (c *K8sClient) GetPodByIP(ip string) (pod *v1.Pod) {
	itf, ok := c.ipIndex.Get(ip)
	if !ok {
		return
	}
	key := itf.(string)
	itf, ok, err := c.Pods().GetStore().GetByKey(key)
	if err != nil {
		log.Error(fmt.Sprintf("get k8s pod[%s] by ip[%s] failed", key, ip), err)
	}
	if !ok {
		return
	}
	pod = itf.(*v1.Pod)
	return
}

func (c *K8sClient) GetNodeByPod(pod *v1.Pod) (node *v1.Node) {
	itf, ok, err := c.Nodes().GetStore().GetByKey(pod.Spec.NodeName)
	if err != nil {
		log.Error(fmt.Sprintf("get k8s node[%s] by pod[%s/%s] failed", pod.Spec.NodeName, pod.Namespace, pod.Name), err)
		return
	}
	if !ok {
		return
	}
	node = itf.(*v1.Node)
	return
}

func (c *K8sClient) Run() {
	if err := c.init(); err != nil {
		err = alarm.Raise(alarm.IDBackendConnectionRefuse,
			alarm.AdditionalContext("%v", err))
		if err != nil {
			log.Error("", err)
		}
		return
	}
	err := alarm.Clear(alarm.IDBackendConnectionRefuse)
	if err != nil {
		log.Error("", err)
	}
	c.goroutine.
		Do(func(_ context.Context) { c.services.Run(c.stopCh) }).
		Do(func(_ context.Context) { c.endpoints.Run(c.stopCh) }).
		Do(func(_ context.Context) { c.pods.Run(c.stopCh) }).
		Do(func(_ context.Context) { c.nodes.Run(c.stopCh) }).
		Do(func(ctx context.Context) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(minWaitInterval):
				util.SafeCloseChan(c.ready)
			}
		})
}

func (c *K8sClient) Stop() {
	close(c.stopCh)
	c.goroutine.Close(true)
	log.Debug("kube client is stopped")
}

func (c *K8sClient) Ready() <-chan struct{} {
	return c.ready
}

func Kubernetes() *K8sClient {
	clientOnce.Do(func() {
		client = &K8sClient{}
		client.Run()
	})
	return client
}

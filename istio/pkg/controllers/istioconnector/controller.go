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

package istioconnector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/istio/pkg/event"
	"github.com/apache/servicecomb-service-center/istio/pkg/utils"
	"github.com/go-chassis/cari/discovery"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/clientset/versioned"
	"istio.io/pkg/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	k8s "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Controller receives service center updates and pushes converted Istio ServiceEntry(s) to k8s api server
type Controller struct {
	// Istio istioClient for k8s API
	istioClient *versioned.Clientset
	// Channel used to send and receive service center change events from the service center controller
	event chan []event.ChangeEvent
	// Lock for service cache
	cacheMutex sync.Mutex
	// Cache of converted service entries, mapped to original service center service id
	convertedServiceCache map[string]*v1alpha3.ServiceEntry
}

func NewController(kubeconfigPath string, e chan []event.ChangeEvent) (*Controller, error) {
	controller := &Controller{
		event:                 e,
		convertedServiceCache: make(map[string]*v1alpha3.ServiceEntry),
	}

	// get kubernetes config info, used for creating k8s client
	client, err := newKubeClient(kubeconfigPath)
	if err != nil {
		log.Errorf("Failed to create istio client: %v\n", err)
		return nil, err
	}
	controller.istioClient = client
	return controller, nil
}

// Return a debounced version of a function `fn` that will not run until `wait` seconds have passed
// after it was last called or until `maxWait` seconds have passed since its first call.
// Once `fn` is executed, the max wait timer is reset.
func debounce(fn func(), wait time.Duration, maxWait time.Duration) func() {
	// Main timer, time seconds elapsed since last execution
	var timer *time.Timer
	// Max wait timer, time seconds elapsed since first call
	var maxTimer *time.Timer
	return func() {
		if maxTimer == nil {
			// First debounced event, start max wait timer
			// will only run target func if not called again after `maxWait` duration
			maxTimer = time.AfterFunc(maxWait, func() {
				// Reset all timers when max wait time is reached
				log.Debugf("Debounce: maximum wait time reached, running target fn\n")
				if timer.Stop() {
					// Only run target func if main timer hasn't already
					fn()
				}
				timer = nil
				maxTimer = nil
			})
			log.Debugf("Debounce: max timer started, will wait max time of %s\n", maxWait)
		}
		if timer != nil {
			// Timer already started; function was called within `wait` duration, debounce this event by resetting timer
			timer.Stop()
		}
		// Start timer, will only run target func if not called again after `wait` duration
		timer = time.AfterFunc(wait, func() {
			log.Debugf("Debounce: timer completed, running target fn\n")
			// Reset all timers and run target func when wait time is reached
			fn()
			maxTimer.Stop()
			maxTimer = nil
			timer = nil
		})
		log.Debugf("Debounce: timer started, will wait %s\n", wait)
	}
}

// Run until a signal is received, this function won't block
func (c *Controller) Run(stop <-chan struct{}) {
	go c.watchServiceCenterUpdate(stop)
}

func (c *Controller) Stop() {

}

// Return a debounced version of the push2istio method that merges the passed events on each call.
func (c *Controller) getIstioPushDebouncer(wait time.Duration, maxWait time.Duration, maxEvents int) func([]event.ChangeEvent) {
	var eventQueue []event.ChangeEvent // Queue of events merged from arguments of each call to debounced function
	// Make a debounced version of push2istio, with provided wait and maxWait times
	debouncedFn := debounce(func() {
		log.Debugf("Debounce: push callback fired, pushing events to Istio: %v\n", eventQueue)
		// Timeout reached, push events to istio and reset queue
		c.push2Istio(eventQueue)
		eventQueue = nil
	}, wait, maxWait)
	return func(newEvents []event.ChangeEvent) {
		log.Debugf("Debounce: received and merged %d new events\n", len(newEvents))
		// Merge new events with existing event queue for each received call
		eventQueue = append(eventQueue, newEvents...)

		log.Debugf("Debounce: new total number of events in queue is %d\n", len(eventQueue))
		if len(eventQueue) > maxEvents {
			c.push2Istio(eventQueue)
			eventQueue = nil
		} else {
			// Make call to debounced push2istio
			debouncedFn()
		}
	}
}

// Watch the Service Center controller for service and instance change events.
func (c *Controller) watchServiceCenterUpdate(stop <-chan struct{}) {
	// Make a debounced push2istio function
	debouncedPush := c.getIstioPushDebouncer(utils.PUSH_DEBOUNCE_INTERVAL, utils.PUSH_DEBOUNCE_MAX_INTERVAL, utils.PUSH_DEBOUNCE_MAX_EVENTS)
	for {
		select {
		case <-stop:
			return
		case events := <-c.event:
			// Received service center change event, use debounced push to Istio.
			// Debouncing introduces latency between the time when change events are received from service center, and when they are pushed to Istio.
			// Debounce latency will take on the range [PUSH_DEBOUNCE_INTERVAL, PUSH_DEBOUNCE_MAX_INTERVAL] in seconds.
			debouncedPush(events)
		}
	}
}

// Push the received service center service/instance change events to Istio.
func (c *Controller) push2Istio(events []event.ChangeEvent) {
	cachedServiceEntries := deepCopyCache(c.convertedServiceCache) // Get cached serviceentries
	for _, e := range events {
		switch ev := e.Event.(type) {
		case *event.MicroserviceEntry:
			// service center service-level change events
			err := c.pushServiceEvent(e.Event.(*event.MicroserviceEntry), e.Action, cachedServiceEntries)
			if err != nil {
				log.Errorf("Failed to push a service center service event to Istio, err[%v]\n", err)
			}
		case *event.InstanceEntry:
			// service center instance-level change events
			err := c.pushEndpointEvents(e.Event.(*event.InstanceEntry), e.Action, cachedServiceEntries)
			if err != nil {
				log.Errorf("Failed to push a service center instance event to Istio, err[%v]\n", err)
			}
		default:
			log.Errorf("Failed to push service center event, event type %T is invalid\n", ev)
		}
	}
	// Save updates to ServiceEntry cache
	c.refreshCache(cachedServiceEntries)
}

// Convert and push service center service-level change events to Istio.
func (c *Controller) pushServiceEvent(e *event.MicroserviceEntry, action discovery.EventType, svcCache map[string]*v1alpha3.ServiceEntry) error {
	serviceId := e.MicroService.ServiceId
	var se *event.ServiceEntry
	// Convert the service center MicroService to an Istio ServiceEntry
	if res := e.Convert(); res == nil {
		return fmt.Errorf("failed to convert service center Service event to Istio ServiceEntry")
	} else {
		se = res.(*event.ServiceEntry)
	}
	name := se.ServiceEntry.GetName()
	log.Debugf("Syncing %s SERVICE event for service center service id %s...\n", string(action), serviceId)
	switch action {
	case discovery.EVT_CREATE:
		// CREATE still requires check to determine whether the service already exists; UPDATE is used in this case.
		// e.g. ServiceEntry fell out of local cache due to controller restart, but in fact already exists in Istio registry.
		fallthrough
	case discovery.EVT_UPDATE:
		existingSe, err := c.istioClient.NetworkingV1alpha3().ServiceEntries(utils.ISTIO_SYSTEM).Get(context.TODO(), name, v1.GetOptions{})
		var returnedSe *v1alpha3.ServiceEntry
		if err != nil {
			returnedSe, err = c.istioClient.NetworkingV1alpha3().ServiceEntries(utils.ISTIO_SYSTEM).Create(context.TODO(), se.ServiceEntry, v1.CreateOptions{})
			if err != nil {
				return err
			}
		} else {
			se.ServiceEntry.Spec.Endpoints = existingSe.Spec.Endpoints // Restore endpoints, only the service itself is being updated
			se.ServiceEntry.Spec.Ports = existingSe.Spec.Ports
			returnedSe, err = c.pushServiceEntryUpdate(existingSe, se.ServiceEntry)
			if err != nil {
				return err
			}
		}
		svcCache[serviceId] = returnedSe
	case discovery.EVT_DELETE:
		err := c.istioClient.NetworkingV1alpha3().ServiceEntries(utils.ISTIO_SYSTEM).Delete(context.TODO(), name, v1.DeleteOptions{})
		if err != nil {
			return err
		}
		delete(svcCache, serviceId)
	}
	log.Infof("Synced %s SERVICE event to Istio\n", string(action))
	return nil
}

// Push an update for an existing ServiceEntry to Istio.
func (c *Controller) pushServiceEntryUpdate(oldServiceEntry, newServiceEntry *v1alpha3.ServiceEntry) (*v1alpha3.ServiceEntry, error) {
	newServiceEntry.SetResourceVersion(oldServiceEntry.GetResourceVersion())
	returnedSe, err := c.istioClient.NetworkingV1alpha3().ServiceEntries(utils.ISTIO_SYSTEM).Update(context.TODO(), newServiceEntry, v1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return returnedSe, nil
}

// Convert and push service center instance-level change events to Istio.
func (c *Controller) pushEndpointEvents(e *event.InstanceEntry, action discovery.EventType, svcCache map[string]*v1alpha3.ServiceEntry) error {
	serviceId := e.ServiceId
	log.Debugf("Syncing %s INSTANCE event for instance %s of service center service %s...\n", string(action), e.InstanceId, serviceId)
	se, ok := svcCache[serviceId]
	if !ok {
		return fmt.Errorf("ServiceEntry for service center Service with id %s was not found", e.ServiceId)
	}
	newSe := se.DeepCopy()
	// Apply changes to the ServiceEntry's endpoints
	err := updateIstioServiceEndpoints(newSe, action, e)
	if err != nil {
		return err
	}
	// Pushed updated ServiceEntry to Istio
	updatedSe, err := c.pushServiceEntryUpdate(se, newSe)
	if err != nil {
		return err
	}
	log.Infof("Pushed %s INSTANCE event to Istio\n", string(action))
	svcCache[serviceId] = updatedSe
	return nil
}

// Apply an update event to a ServiceEntry's endpoint(s).
func updateIstioServiceEndpoints(se *v1alpha3.ServiceEntry, action discovery.EventType, targetInst *event.InstanceEntry) error {
	targetInstanceId := targetInst.InstanceId
	newInsts := []*event.InstanceEntry{}
	var seAsMSE *event.MicroserviceEntry
	// Convert ServiceEntry back to service center service to apply changes to its service center instances
	if res := event.NewServiceEntry(se).Convert(); res == nil {
		return fmt.Errorf("failed to parse existing Istio ServiceEntry")
	} else {
		seAsMSE = res.(*event.MicroserviceEntry)
	}
	switch discovery.EventType(action) {
	case discovery.EVT_DELETE:
		// Filter out the deleted instance
		for _, existingInst := range seAsMSE.Instances {
			if existingInst.InstanceId != targetInstanceId {
				newInsts = append(newInsts, existingInst)
			}
		}
		if len(seAsMSE.Instances) == len(newInsts) {
			log.Warnf("could not push delete for target Service Center instance id %s, instance was not found\n", targetInstanceId)
		}
		seAsMSE.Instances = newInsts
	case discovery.EVT_CREATE:
		// CREATE still requires check to determine whether the endpoint already exists; UPDATE is used in this case.
		fallthrough
	case discovery.EVT_UPDATE:
		updated := false
		for i, existingInst := range seAsMSE.Instances {
			if existingInst.InstanceId == targetInstanceId {
				// Found existing instance, update with new instance
				seAsMSE.Instances[i] = targetInst
				updated = true
				break
			}
		}
		if !updated {
			// Instance does not already exist, add as new instance
			seAsMSE.Instances = append(seAsMSE.Instances, targetInst)
		}
	}
	// Convert the microservice entry back to istio service entry; the serviceports for the changed endpoints will be regenerated appropriately by conversion logic
	var regenedSe *v1alpha3.ServiceEntry
	if res := seAsMSE.Convert(); res == nil {
		return fmt.Errorf("failed to parse changes for Istio ServiceEntry")
	} else {
		regenedSe = res.(*event.ServiceEntry).ServiceEntry
	}
	// Only take regened ports and new workloadentries, preserves rest of original serviceentry
	se.Spec.Endpoints = regenedSe.Spec.Endpoints
	se.Spec.Ports = regenedSe.Spec.Ports
	return nil
}

// Save Istio ServiceEntry(s) converted from service center updates.
func (c *Controller) refreshCache(serviceEntries map[string]*v1alpha3.ServiceEntry) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.convertedServiceCache = serviceEntries
}

// Get a deep copy of the converted Istio ServiceEntry(s) pushed from service center.
func deepCopyCache(m map[string]*v1alpha3.ServiceEntry) map[string]*v1alpha3.ServiceEntry {
	newMap := make(map[string]*v1alpha3.ServiceEntry)
	for k, v := range m {
		newMap[k] = v.DeepCopy()
	}
	return newMap
}

// newKubeClient creates new kube client
func newKubeClient(kubeconfigPath string) (*versioned.Clientset, error) {
	var err error
	var kubeConf *rest.Config

	if kubeconfigPath == "" {
		// creates the in-cluster config
		kubeConf, err = k8s.GetConfig()
		if err != nil {
			return nil, fmt.Errorf("build default in cluster kube config failed: %w", err)
		}
	} else {
		kubeConf, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("build kube client config from config file failed: %w", err)
		}
	}
	return versioned.NewForConfig(kubeConf)
}

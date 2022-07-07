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

package bootstrap

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/apache/servicecomb-service-center/istio/pkg/controllers/istioconnector"
	"github.com/apache/servicecomb-service-center/istio/pkg/controllers/servicecenter"
	"github.com/apache/servicecomb-service-center/istio/pkg/event"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"istio.io/pkg/log"
)

// cli args
type Args struct {
	// servicecomb-service-center address
	ServiceCenterAddr string
	// kubeconfig file path
	Kubeconfig string
	// enable leader election or not for high abalibility
	HA bool
}

const (
	// leader election check locked resource namespace
	lockNameSpace = "istio-system"
	// leader election check locked resource name
	resourceName = "servicecenter2mesh"
	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack.
	//
	// A client needs to wait a full LeaseDuration without observing a change to
	// the record before it can attempt to take over. When all clients are
	// shutdown and a new set of clients are started with different names against
	// the same leader record, they must wait the full LeaseDuration before
	// attempting to acquire the lease. Thus LeaseDuration should be as short as
	// possible (within your tolerance for clock skew rate) to avoid a possible
	// long waits in the scenario.
	//
	// Core clients default this value to 15 seconds.
	defaultLeaseDuration = 15 * time.Second
	// RenewDeadline is the duration that the acting master will retry
	// refreshing leadership before giving up.
	//
	// Core clients default this value to 10 seconds.
	defaultRenewDeadline = 10 * time.Second
	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	//
	// Core clients default this value to 2 seconds.
	defaultRetryPeriod = 2 * time.Second
)

type Server struct {
	// service center controller watches service center update, and push to istio controller
	serviceCenterController *servicecenter.Controller
	// istio controller receives updates from service center controller and push to k8s api server
	istioController *istioconnector.Controller
	// channel for passing service center event from service center controller to istio controller
	serviceCenterEvent chan []event.ChangeEvent
}

func NewServer(args *Args) (*Server, error) {
	// only allow 1 eventlist at a time
	changeEvent := make(chan []event.ChangeEvent, 1)

	// Create a new istio controller, the controller is ready to push configs to istio
	istioController, err := istioconnector.NewController(args.Kubeconfig, changeEvent)
	if err != nil {
		return nil, err
	}

	serviceCenterController := servicecenter.NewController(args.ServiceCenterAddr, changeEvent)

	s := &Server{
		serviceCenterController: serviceCenterController,
		istioController:         istioController,
		serviceCenterEvent:      changeEvent,
	}

	return s, nil
}

// start the server need to start both service center and istio controller
func (s *Server) Start(ctx context.Context, args *Args) error {
	// by default the leader election is disabled, just do regular start
	if !args.HA {
		s.doRun(ctx)
		return nil
	}

	return s.doLeaderElectionRun(ctx)
}

// This function is used to enable leader election using k8s client-go api. leaderElectAndRun runs the leader election,
// and runs the callbacks once the leader lease is acquired.
//
// For k8s clint-go API:
// DISCLAIMER: this is an alpha API. This library will likely change significantly or even be removed entirely in subsequent releases.
// Depend on this API at your own risk.
//
// Note: this API is also used by K8S controller and Cluster auto scaler.
func (s *Server) doLeaderElectionRun(ctx context.Context) error {
	id, err := os.Hostname()
	if err != nil {
		return err
	}

	// creates the in-cluster config
	kubeConf, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("build default in cluster kube config failed: %w", err)
	}

	client, err := kubernetes.NewForConfig(kubeConf)

	if err != nil {
		log.Fatalf("build kube client failed: %v", err)
		return err
	}

	rl, err := resourcelock.New(resourcelock.LeasesResourceLock,
		lockNameSpace,
		resourceName,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: id,
		})
	if err != nil {
		log.Fatalf("error creating lock: %v", err)
		return err
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            rl,
		ReleaseOnCancel: true,
		LeaseDuration:   defaultLeaseDuration,
		RenewDeadline:   defaultRenewDeadline,
		RetryPeriod:     defaultRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) {
				s.doRun(ctx)
			},
			OnStoppedLeading: func() {
				log.Infof("%s: stopped leading", id)
			},
		},
	})
	return nil
}

func (s *Server) doRun(ctx context.Context) {
	go s.serviceCenterController.Run(ctx)
	go s.istioController.Run(ctx)
	log.Info("servicecenter2mesh Server Started !!!")

	s.waitForShutdown(ctx)
}

// on server stop
func (s *Server) waitForShutdown(ctx context.Context) {
	go func() {
		<-ctx.Done()
		s.serviceCenterController.Stop()
		close(s.serviceCenterEvent)
	}()
}

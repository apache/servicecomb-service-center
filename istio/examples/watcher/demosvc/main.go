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

package main

import (
	"fmt"
	"os"

	scregistry "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/go-chassis/v2/core/common"
	"github.com/go-chassis/sc-client"
)

func main() {
	printDemoMsg("--- DEMO APP STARTED ---\n\t- The following steps will demonstrate the Watcher features of servicecenter2mesh.\n\t- To check the changes in Istio, we must make cURL requests to Istio API's ServiceEntry endpoint.")

	breakpoint()

	registryClient, err := sc.NewClient(
		sc.Options{
			Endpoints: []string{"127.0.0.1:30100"},
		})
	if err != nil {
		fmt.Printf("err[%v]\n", err)
		os.Exit(1)
	}

	service := &scregistry.MicroService{
		AppId:       "demoapp",
		ServiceName: "myserver1",
		Version:     "0.0.1",
		Environment: "",
	}
	printDemoMsg(fmt.Sprintf("1. We register a new Service Center MicroService with name[%s], appId[%s]", service.ServiceName, service.AppId))
	sid, err := registryClient.RegisterService(service)
	if err != nil {
		fmt.Printf("err[%v]\n", err)
		os.Exit(1)
	}
	fmt.Printf("RegisterService %s with sid[%s]\n", service.ServiceName, sid)

	instance := scregistry.MicroServiceInstance{
		ServiceId: sid,
		HostName:  "demohost1",
		Status:    common.DefaultStatus,
		Endpoints: []string{"rest://1.1.1.1:8888"},
	}
	printDemoMsg(fmt.Sprintf("2. We register a new Service Center MicroServiceInstance for service with name[%s]", service.ServiceName))
	iid, err := registryClient.RegisterMicroServiceInstance(&instance)
	if err != nil {
		fmt.Printf("RegisterMicroServiceInstance, err[%v]\n", err)
		os.Exit(1)
	}
	fmt.Printf("RegisterMicroServiceInstance with iid[%s] for service with sid[%s]\n", iid, sid)

	printServiceCenter2MeshMsg(fmt.Sprintf("We are expecting:\n\t- Full sync triggered and synced new service with name[%s] to Istio: ServiceEntry created for service with name[%s], with one WorkloadEntry endpoint\n\t- Watcher service created for appId[%s], that will start to watch for new instance events", service.ServiceName, service.ServiceName, service.AppId))

	breakpoint()

	printDemoMsg(fmt.Sprintf("3. We unregister the instance for service with name[%s]", service.ServiceName))
	ok, err := registryClient.UnregisterMicroServiceInstance(sid, iid)
	if !ok {
		fmt.Printf("UnregisterMicroServiceInstance, err[%v]\n", err)
		os.Exit(1)
	}
	fmt.Printf("UnregisterMicroServiceInstance with iid[%s] for service with sid[%s]\n", iid, sid)

	printServiceCenter2MeshMsg(fmt.Sprintf("We are expecting:\n\t- DELETE instance event detected by instance watcher: corresponding WorkloadEntry deleted from endpoints of ServiceEntry with name[%s]", service.ServiceName))

	breakpoint()

	service2 := &scregistry.MicroService{
		AppId:       "demoapp",
		ServiceName: "myserver2",
		Version:     "0.0.1",
		Environment: "",
	}

	printDemoMsg(fmt.Sprintf("4. We register a new Service Center MicroService with name %s, same appId[%s] as %s", service2.ServiceName, service2.AppId, service.ServiceName))

	sid2, err := registryClient.RegisterService(service2)
	if err != nil {
		fmt.Printf("err[%v]\n", err)
		os.Exit(1)
	}
	fmt.Printf("RegisterService %s with sid[%s]\n", service2.ServiceName, sid2)

	instance2 := scregistry.MicroServiceInstance{
		ServiceId: sid2,
		HostName:  "demohost2",
		Status:    common.DefaultStatus,
		Endpoints: []string{"rest://2.2.2.2:8888"},
	}
	printDemoMsg(fmt.Sprintf("5. We register a new Service Center MicroServiceInstance for service2 with name[%s]", service2.ServiceName))
	iid2, err := registryClient.RegisterMicroServiceInstance(&instance2)
	if err != nil {
		fmt.Printf("RegisterMicroServiceInstance, err[%v]\n", err)
		os.Exit(1)
	}
	fmt.Printf("RegisterMicroServiceInstance with iid[%s] for service with sid[%s]\n", iid2, sid2)
	printServiceCenter2MeshMsg(fmt.Sprintf("We are expecting:\n\t- Since watcher can only detect INSTANCE changes, full sync is used to sync new service with name[%s] to Istio: ServiceEntry created for service with name[%s], with one WorkloadEntry endpoint\n\t- Previous instance watcher already exists for same appId[%s], so no new watcher needed", service2.ServiceName, service2.ServiceName, service2.AppId))

	breakpoint()

	printDemoMsg(fmt.Sprintf("6. We unregister the instance for service with name[%s]", service2.ServiceName))
	ok, err = registryClient.UnregisterMicroServiceInstance(sid2, iid2)
	if !ok {
		fmt.Printf("UnregisterMicroServiceInstance, err[%v]\n", err)
		os.Exit(1)
	}
	fmt.Printf("UnregisterMicroServiceInstance with iid[%s] for service with sid[%s]\n", iid2, sid2)

	printServiceCenter2MeshMsg(fmt.Sprintf("We are expecting:\n\t- DELETE instance event detected by existing instance watcher for appId[%s]: corresponding WorkloadEntry deleted from endpoints of ServiceEntry with name[%s]", service2.AppId, service2.ServiceName))

	breakpoint()

	printDemoMsg(fmt.Sprintf("7. We unregister both service with name[%s] and service with name[%s]", service.ServiceName, service2.ServiceName))
	ok, err = registryClient.UnregisterMicroService(sid)
	if !ok {
		fmt.Printf("UnregisterMicroService, err[%v]\n", err)
		os.Exit(1)
	}
	fmt.Printf("UnregisterMicroService for service with sid[%s]\n", sid)
	ok, err = registryClient.UnregisterMicroService(sid2)
	if !ok {
		fmt.Printf("UnregisterMicroService, err[%v]\n", err)
		os.Exit(1)
	}
	fmt.Printf("UnregisterMicroService for service with sid[%s]\n", sid2)

	printServiceCenter2MeshMsg("We are expecting:\n\t- Full sync DELETE SERVICE for both services to Istio, both ServiceEntries removed from Istio")

	breakpoint()

	printDemoMsg("--- DEMO COMPLETE ---")
}

func breakpoint() {
	fmt.Printf("\n[BREAKPOINT] Execution Paused\n")
	fmt.Scanln()
}

func printDemoMsg(msg string) {
	fmt.Printf("\n[DEMO] %s\n", msg)
}

func printServiceCenter2MeshMsg(msg string) {
	fmt.Printf("\n[servicecenter2mesh] %s\n", msg)
}

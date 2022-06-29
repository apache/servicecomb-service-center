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
	"time"

	scregistry "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/go-chassis/v2/core/common"
	"github.com/go-chassis/sc-client"
)

func main() {
	printDemoMsg("--- DEMO APP STARTED ---\n\t- The following steps will demonstrate the Debounce features of servicecenter2mesh.\n\t- To check the changes in Istio, we must make cURL requests to Istio API's ServiceEntry endpoint.")

	breakpoint()

	registryClient, err := sc.NewClient(
		sc.Options{
			Endpoints: []string{"127.0.0.1:30100"},
		})
	if err != nil {
		fmt.Printf("err[%v]\n", err)
		os.Exit(1)
	}

	appId := "demoapp"
	service := &scregistry.MicroService{
		AppId:       appId,
		Version:     "0.0.1",
		ServiceName: "myserver1",
		Environment: "",
	}
	printDemoMsg(fmt.Sprintf("1. We will register a new service center MicroService with name[%s] and appId[%s]", service.ServiceName, appId))
	sid, err := registryClient.RegisterService(service)
	if err != nil {
		fmt.Printf("err[%v]\n", err)
		os.Exit(1)
	}
	fmt.Printf("RegisterService %s with sid[%s]\n", service.ServiceName, sid)
	breakpoint()

	printDemoMsg("2. We register a new instance of the service, every 1 second, for 10 seconds")
	var iids []string
	for i := 0; i < 10; i++ {
		instance := &scregistry.MicroServiceInstance{
			ServiceId: sid,
			HostName:  fmt.Sprintf("demohost%d", i),
			Status:    common.DefaultStatus,
			Endpoints: []string{fmt.Sprintf("rest://%d.%d.%d.%d:8888", i, i, i, i)},
		}
		fmt.Printf("Registering instance: %v\n", *instance)
		iid, err := registryClient.RegisterMicroServiceInstance(instance)
		if err != nil {
			fmt.Printf("err[%v]\n", err)
			os.Exit(1)
		}
		iids = append(iids, iid)
		time.Sleep(time.Second * 1)
	}
	printServiceCenter2MeshMsg("Max timer of 10 secs will expire, will trigger istio push will all events merged into single array")
	breakpoint()

	printDemoMsg("2. We unregister all instances, 2 at a time with a breakpoint in between")
	for i := 0; i < 10; i += 2 {
		_, err := registryClient.UnregisterMicroServiceInstance(sid, iids[i])
		if err != nil {
			fmt.Printf("err[%v]\n", err)
			os.Exit(1)
		}
		fmt.Printf("UnregisterMicroServiceInstance with iid[%s]\n", iids[i])
		_, err = registryClient.UnregisterMicroServiceInstance(sid, iids[i+1])
		if err != nil {
			fmt.Printf("err[%v]\n", err)
			os.Exit(1)
		}
		fmt.Printf("UnregisterMicroServiceInstance with iid[%s]\n", iids[i+1])
		breakpoint()
	}
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

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

package metric

import (
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/astaxie/beego"
	"net"
	"os"
	"sync"
	"time"
)

const (
	defaultCollectPeriod = 30 * time.Second
	FamilyName           = "service_center"
	familyNamePrefix     = FamilyName + "_"
	bufferSize           = 1024
)

var (
	// Period is metrics collect period
	Period = 30 * time.Second
	// SysMetrics map
	SysMetrics util.ConcurrentMap

	getEndpointOnce sync.Once
	instance        string
)

func init() {
	Period = getPeriod()
	SysMetrics.Put("process_resident_memory_bytes", struct{}{})
	SysMetrics.Put("process_cpu_seconds_total", struct{}{})
	SysMetrics.Put("go_threads", struct{}{})
	SysMetrics.Put("go_goroutines", struct{}{})
}

func getPeriod() time.Duration {
	inv := os.Getenv("METRICS_INTERVAL")
	d, err := time.ParseDuration(inv)
	if err == nil && d >= time.Second {
		return d
	}
	return defaultCollectPeriod
}

func InstanceName() string {
	getEndpointOnce.Do(func() {
		restIP := beego.AppConfig.String("httpaddr")
		restPort := beego.AppConfig.String("httpport")
		if len(restIP) > 0 {
			instance = net.JoinHostPort(restIP, restPort)
			return
		}

		rpcIP := beego.AppConfig.String("rpcaddr")
		rpcPort := beego.AppConfig.String("rpcport")
		if len(rpcIP) > 0 {
			instance = net.JoinHostPort(rpcIP, rpcPort)
			return
		}
	})
	return instance
}

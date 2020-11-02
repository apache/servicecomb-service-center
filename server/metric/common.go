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
)

const (
	FamilyName       = "service_center"
	familyNamePrefix = FamilyName + "_"
	bufferSize       = 1024
)

var (
	options Options
	// SysMetrics map
	SysMetrics util.ConcurrentMap
	// Gatherer is the reader of sc metrics
	Gatherer *MetricsGatherer
)

func Init(opts Options) error {
	options = opts
	for _, key := range options.SysMetrics {
		SysMetrics.Put(key, struct{}{})
	}
	Gatherer = NewGatherer(opts)
	Gatherer.Start()
	return nil
}

func GetOptions() Options {
	return options
}

func InstanceName() string {
	return GetOptions().InstanceName
}

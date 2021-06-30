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

package metrics

import (
	metricsvc "github.com/apache/servicecomb-service-center/pkg/metrics"
	"github.com/go-chassis/go-chassis/v2/pkg/metrics"
)

func Init(options Options) error {
	if err := metrics.Init(); err != nil {
		return err
	}
	if err := metricsvc.Init(metricsvc.Options{
		Interval:     options.Interval,
		InstanceName: options.Instance,
		SysMetrics: []string{
			"process_resident_memory_bytes",
			"process_cpu_seconds_total",
			"go_threads",
			"go_goroutines",
		},
	}); err != nil {
		return err
	}

	if err := InitMetaMetrics(); err != nil {
		return err
	}
	return nil
}

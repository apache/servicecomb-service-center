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

package health

import (
	"context"
	"runtime"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const durationReportCPUUsage = 3 * time.Second

var (
	cpuGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: "process",
			Name:      "cpu_usage",
			Help:      "Process cpu usage",
		}, []string{"instance"})
)

func init() {
	prometheus.MustRegister(cpuGauge)
	gopool.Go(AutoReportCPUUsage)
}

func AutoReportCPUUsage(ctx context.Context) {
	var (
		cpuTotal float64
		cpuProc  float64
		cpus     = runtime.NumCPU()
	)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(durationReportCPUUsage):
			pt, ct := util.GetProcCPUUsage()
			cpuGauge.WithLabelValues(metrics.InstanceName()).Set(
				(pt - cpuProc) * float64(cpus) / (ct - cpuTotal))
			cpuTotal, cpuProc = ct, pt
		}
	}
}

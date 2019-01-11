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
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/server/metric"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
	"golang.org/x/net/context"
	"os"
	"time"
)

var (
	cpu = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: "process",
			Name:      "cpu_usage",
			Help:      "Process cpu usage",
		}, []string{"instance"})
)

func init() {
	prometheus.MustRegister(cpu)
	gopool.Go(func(ctx context.Context) {
		var (
			cpuTotal float64
			cpuProc  float64
		)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(3 * time.Second):
				p, _ := procfs.NewProc(os.Getpid())
				stat, _ := procfs.NewStat()
				pstat, _ := p.NewStat()
				ct := stat.CPUTotal.User + stat.CPUTotal.Nice + stat.CPUTotal.System +
					stat.CPUTotal.Idle + stat.CPUTotal.Iowait + stat.CPUTotal.IRQ +
					stat.CPUTotal.SoftIRQ + stat.CPUTotal.Steal + stat.CPUTotal.Guest
				pt := float64(pstat.UTime+pstat.STime+pstat.CUTime+pstat.CSTime) / 100
				cpu.WithLabelValues(metric.InstanceName()).Set(
					(pt - cpuProc) * float64(len(stat.CPU)) / (ct - cpuTotal))
				cpuTotal, cpuProc = ct, pt
			}

		}
	})
}

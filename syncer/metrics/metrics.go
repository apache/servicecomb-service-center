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
	"net"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	metricsvc "github.com/apache/servicecomb-service-center/pkg/metrics"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/go-chassis/go-chassis/v2/pkg/metrics"
)

const (
	FamilyName        = "syncer"
	KeyPendingEvent   = FamilyName + "_pending_event"
	KeyAbandonEvent   = FamilyName + "_abandon_event"
	KeyPendingTask    = FamilyName + "_pending_task"
	KeyConnectedPeers = FamilyName + "_connected_peers"
	KeyPeersTotal     = FamilyName + "_peers_total"
	KeyPeersClockDiff = FamilyName + "_peers_clock_diff"
)

var Instance string

func Init() error {
	if !config.GetBool("metrics.enable", false) {
		return nil
	}

	Instance = net.JoinHostPort(config.GetString("server.host", "", config.WithStandby("httpaddr")),
		config.GetString("server.port", "", config.WithStandby("httpport")))

	metricsvc.CollectFamily(FamilyName)
	//TODO should call metrics.Init()

	if err := metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeyPendingEvent,
		Help:   "The number of events pending to send",
		Labels: []string{"instance"},
	}); err != nil {
		return err
	}

	if err := metrics.CreateCounter(metrics.CounterOpts{
		Key:    KeyAbandonEvent,
		Help:   "The number of abandon events",
		Labels: []string{"instance"},
	}); err != nil {
		return err
	}

	if err := metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeyPendingTask,
		Help:   "The number of tasks pending to dispatch",
		Labels: []string{"instance"},
	}); err != nil {
		return err
	}

	if err := metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeyConnectedPeers,
		Help:   "The number of connected peers",
		Labels: []string{"instance"},
	}); err != nil {
		return err
	}

	if err := metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeyPeersTotal,
		Help:   "The number of peers",
		Labels: []string{"instance"},
	}); err != nil {
		return err
	}

	if err := metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeyPeersClockDiff,
		Help:   "The diff milliseconds of peers",
		Labels: []string{"instance", "peer"},
	}); err != nil {
		return err
	}
	return nil
}

func PendingEventSet(n int64) {
	labels := map[string]string{"instance": Instance}
	if err := metrics.GaugeSet(KeyPendingEvent, float64(n), labels); err != nil {
		log.Error("gauge set failed", err)
	}
}

func AbandonEventAdd() {
	labels := map[string]string{"instance": Instance}
	if err := metrics.CounterAdd(KeyAbandonEvent, 1, labels); err != nil {
		log.Error("counter add failed", err)
	}
}

func PendingTaskSet(n int64) {
	labels := map[string]string{"instance": Instance}
	if err := metrics.GaugeSet(KeyPendingTask, float64(n), labels); err != nil {
		log.Error("gauge set failed", err)
	}
}

func ConnectedPeersSet(n int64) {
	labels := map[string]string{"instance": Instance}
	if err := metrics.GaugeSet(KeyConnectedPeers, float64(n), labels); err != nil {
		log.Error("gauge set failed", err)
	}
}

func PeersTotalSet(n int64) {
	labels := map[string]string{"instance": Instance}
	if err := metrics.GaugeSet(KeyPeersTotal, float64(n), labels); err != nil {
		log.Error("gauge set failed", err)
	}
}

func PeersClockDiffSet(peerName string, n int64) {
	labels := map[string]string{"instance": Instance, "peer": peerName}
	if err := metrics.GaugeSet(KeyPeersClockDiff, float64(n/time.Millisecond.Nanoseconds()), labels); err != nil {
		log.Error("gauge set failed", err)
	}
}

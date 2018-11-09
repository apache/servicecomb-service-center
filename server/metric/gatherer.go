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
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pm "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"strings"
	"sync"
	"time"
)

var Gatherer *MetricsGatherer

func init() {
	Gatherer = NewGatherer()
	Gatherer.Start()
}

func NewGatherer() *MetricsGatherer {
	return &MetricsGatherer{
		Records: NewMetrics(),
		closed:  true,
	}
}

type MetricsGatherer struct {
	Records *Metrics

	lock   sync.Mutex
	closed bool
}

func (mm *MetricsGatherer) Start() {
	mm.lock.Lock()
	if !mm.closed {
		mm.lock.Unlock()
		return
	}
	mm.closed = false

	gopool.Go(mm.loop)

	mm.lock.Unlock()
}

func (mm *MetricsGatherer) loop(ctx context.Context) {
	ticker := time.NewTicker(Period)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := mm.Collect(); err != nil {
				log.Errorf(err, "metrics collect failed")
				return
			}

			Report()
		}
	}
}

func (mm *MetricsGatherer) Collect() error {
	mfs, err := pm.DefaultGatherer.Gather()
	if err != nil {
		return err
	}

	for _, mf := range mfs {
		name := mf.GetName()
		if _, ok := SysMetrics.Get(name); strings.Index(name, familyNamePrefix) == 0 || ok {
			if d := Calculate(mf); d != nil {
				mm.Records.Put(strings.TrimPrefix(name, familyNamePrefix), d)
			}
		}
	}
	return nil
}

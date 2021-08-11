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

package etcd

import (
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/metrics"
)

func TestReportCacheSize(t *testing.T) {
	if err := metrics.Init(metrics.Options{
		Interval:     time.Second,
		InstanceName: "test",
	}); err != nil {
		t.Fatalf("init metric failed %v", err)
	}
	ReportCacheSize("a", "b", 100)
	err := metrics.Gatherer.Collect()
	if err != nil {
		t.Fatalf("TestReportCacheSize failed")
	}
	if metrics.Gatherer.Records.Summary("local_cache_size_bytes") != 100 {
		t.Fatalf("TestReportCacheSize failed")
	}

	ReportCacheSize("", "b", 200)
	err = metrics.Gatherer.Collect()
	if metrics.Gatherer.Records.Summary("local_cache_size_bytes") != 100 {
		t.Fatalf("TestReportCacheSize failed")
	}
}

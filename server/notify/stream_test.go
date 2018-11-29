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
package notify

import (
	"github.com/apache/servicecomb-service-center/pkg/log"
	"golang.org/x/net/context"
	"testing"
)

func TestHandleWatchJob(t *testing.T) {
	defer log.Recover()
	w := NewInstanceEventListWatcher("g", "s", nil)
	w.Job <- nil
	err := HandleWatchJob(w, nil)
	if err == nil {
		t.Fatalf("TestHandleWatchJob failed")
	}
	w.Job <- NewInstanceEvent("g", "s", 1, nil)
	err = HandleWatchJob(w, nil)
	t.Fatalf("TestHandleWatchJob failed")
}

func TestDoStreamListAndWatch(t *testing.T) {
	err := DoStreamListAndWatch(context.Background(), "s", nil, nil)
	if err == nil {
		t.Fatal("TestDoStreamListAndWatch failed", err)
	}
}

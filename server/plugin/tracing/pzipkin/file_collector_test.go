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
package pzipkin

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
	"os"
	"testing"
	"time"
)

func TestFileCollector_Collect(t *testing.T) {
	bad := os.NewFile(0, "/dev/null") // bad fd
	fc := &FileCollector{
		Fd:        bad,
		Timeout:   2 * time.Second,
		Interval:  1 * time.Second,
		BatchSize: 2,
		c:         make(chan *zipkincore.Span, 100),
		goroutine: gopool.New(context.Background()),
	}
	fc.Run()

	// 0, 1 drop [2, 3, 4, 5], [6, 7] ok, 8 delay ok
	for i := 0; i < 9; i++ {
		if i > 4 {
			fc.Fd = os.Stdout
		}
		err := fc.Collect(&zipkincore.Span{ID: int64(i)})
		if err != nil {
			t.Fatalf("TestFileCollector_Collect failed, %s", err.Error())
		}
		<-time.After(100 * time.Millisecond)
		fmt.Println(i)
	}

	<-time.After(1 * time.Second)

	// fc.Close()
}

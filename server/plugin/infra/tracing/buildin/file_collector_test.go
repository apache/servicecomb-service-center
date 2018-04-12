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
package buildin

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
	"golang.org/x/net/context"
	"os"
	"testing"
	"time"
)

func TestFileCollector_Collect(t *testing.T) {
	fc := &FileCollector{
		Fd:        os.Stdout,
		Timeout:   1 * time.Second,
		Interval:  100 * time.Second,
		BatchSize: 2,
		c:         make(chan *zipkincore.Span, 100),
		goroutine: util.NewGo(context.Background()),
	}
	defer func() {
		fc.Close()
	}()
	fc.Run()

	for i := 0; i < 10; i++ {
		err := fc.Collect(&zipkincore.Span{})
		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}
	}

	<-time.After(time.Second)
}

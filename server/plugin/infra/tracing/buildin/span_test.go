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
	"encoding/json"
	"fmt"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
	"testing"
)

var sample = []byte(`{
  "trace_id": 4081433150731846551,
  "name": "api",
  "id": 8350480249446290292,
  "annotations": [
    {
      "timestamp": 1517455386250894,
      "value": "sr",
      "host": {
        "ipv4": 2130706433,
        "port": 30100,
        "service_name": "SERVICECENTER"
      }
    },
    {
      "timestamp": 1517455386251872,
      "value": "ss",
      "host": {
        "ipv4": 2130706433,
        "port": 30100,
        "service_name": "SERVICECENTER"
      }
    }
  ],
  "binary_annotations": [
    {
      "key": "span.kind",
      "value": "c2VydmVy",
      "annotation_type": "STRING",
      "host": {
        "ipv4": 2130706433,
        "port": 30100,
        "service_name": "SERVICECENTER"
      }
    },
    {
      "key": "protocol",
      "value": "SFRUUA==",
      "annotation_type": "STRING",
      "host": {
        "ipv4": 2130706433,
        "port": 30100,
        "service_name": "SERVICECENTER"
      }
    },
    {
      "key": "http.method",
      "value": "R0VU",
      "annotation_type": "STRING",
      "host": {
        "ipv4": 2130706433,
        "port": 30100,
        "service_name": "SERVICECENTER"
      }
    },
    {
      "key": "http.url",
      "value": "L3Y0L2RlZmF1bHQvcmVnaXN0cnkvbWljcm9zZXJ2aWNlcz8lM0Fwcm9qZWN0PWRlZmF1bHQm",
      "annotation_type": "STRING",
      "host": {
        "ipv4": 2130706433,
        "port": 30100,
        "service_name": "SERVICECENTER"
      }
    },
    {
      "key": "http.path",
      "value": "L3Y0L2RlZmF1bHQvcmVnaXN0cnkvbWljcm9zZXJ2aWNlcw==",
      "annotation_type": "STRING",
      "host": {
        "ipv4": 2130706433,
        "port": 30100,
        "service_name": "SERVICECENTER"
      }
    },
    {
      "key": "http.host",
      "value": "",
      "annotation_type": "STRING",
      "host": {
        "ipv4": 2130706433,
        "port": 30100,
        "service_name": "SERVICECENTER"
      }
    },
    {
      "key": "resultCode",
      "value": "MjAw",
      "annotation_type": "STRING",
      "host": {
        "ipv4": 2130706433,
        "port": 30100,
        "service_name": "SERVICECENTER"
      }
    },
    {
      "key": "result",
      "value": "MA==",
      "annotation_type": "STRING",
      "host": {
        "ipv4": 2130706433,
        "port": 30100,
        "service_name": "SERVICECENTER"
      }
    },
    {
      "key": "http.status_code",
      "value": "MjAw",
      "annotation_type": "STRING",
      "host": {
        "ipv4": 2130706433,
        "port": 30100,
        "service_name": "SERVICECENTER"
      }
    }
  ],
  "timestamp": 1517455386250894,
  "duration": 978,
  "trace_id_high": 7511721612091346172
}`)

func TestFromZipkinSpan(t *testing.T) {
	span := &zipkincore.Span{}
	err := json.Unmarshal(sample, &span)
	if err != nil {
		fmt.Println("TestFromZipkinSpan Unmarshal", err)
		t.FailNow()
	}
	s := FromZipkinSpan(span)
	b, err := json.Marshal(s)
	if err != nil {
		fmt.Println("TestFromZipkinSpan Marshal", err)
		t.FailNow()
	}
	fmt.Println(string(b))
}

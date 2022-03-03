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
	"os"
	"path/filepath"
	"testing"

	zipkintracer "github.com/openzipkin/zipkin-go-opentracing"

	"github.com/go-chassis/go-archaius"
)

func TestGetFilePath(t *testing.T) {
	wd, _ := os.Getwd()
	f := GetFilePath("a")
	if f != filepath.Join(wd, "a") {
		t.Fatalf("TestGetFilePath failed, %v", f)
	}
	archaius.Set(fileCollectorPath, "trace.log")
	f = GetFilePath("a")
	if f != "trace.log" {
		t.Fatalf("TestGetFilePath failed, %v", f)
	}
}

func TestGetSamplerRate(t *testing.T) {
	r := GetSamplerRate()
	if r != defaultSamplerRate {
		t.Fatalf("TestGetSamplerRate failed, %v", r)
	}
	archaius.Set(samplerRate, "a")
	r = GetSamplerRate()
	if r != defaultSamplerRate {
		t.Fatalf("TestGetSamplerRate failed, %v", r)
	}
	archaius.Set(samplerRate, "0.1")
	r = GetSamplerRate()
	if r != 0.1 {
		t.Fatalf("TestGetSamplerRate failed, %v", r)
	}
}

func TestNewCollector(t *testing.T) {
	archaius.Set(collectorType, "")
	tracer, err := newCollector()
	if err == nil {
		t.Fatalf("TestNewCollector failed")
	}
	archaius.Set(collectorType, "server")
	tracer, err = newCollector()
	if err != nil {
		t.Fatalf("TestNewCollector failed")
	}
	_, ok := tracer.(*zipkintracer.HTTPCollector)
	if !ok {
		t.Fatalf("TestNewCollector failed")
	}
	archaius.Set(collectorType, "file")
	tracer, err = newCollector()
	if err != nil {
		t.Fatalf("TestNewCollector failed")
	}
	_, ok = tracer.(*FileCollector)
	if !ok {
		t.Fatalf("TestNewCollector failed")
	}
}

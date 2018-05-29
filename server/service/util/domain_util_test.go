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
package util

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"golang.org/x/net/context"
	"testing"
)

func TestGetDomain(t *testing.T) {
	_, err := GetAllDomainRawData(util.SetContext(context.Background(), CTX_CACHEONLY, "1"))
	if err != nil {
		t.Fatalf("GetAllDomainRawData WithCacheOnly failed")
	}

	_, err = GetAllDomainRawData(context.Background())
	if err == nil {
		t.Fatalf("GetAllDomainRawData failed")
	}

	_, err = GetAllDomain(util.SetContext(context.Background(), CTX_CACHEONLY, "1"))
	if err != nil {
		t.Fatalf("GetAllDomain WithCacheOnly failed")
	}

	_, err = GetAllDomain(context.Background())
	if err == nil {
		t.Fatalf("GetAllDomain failed")
	}
}

func TestDomainExist(t *testing.T) {
	_, err := DomainExist(util.SetContext(context.Background(), CTX_CACHEONLY, "1"), "")
	if err != nil {
		t.Fatalf("DomainExist WithCacheOnly failed")
	}

	_, err = DomainExist(context.Background(), "")
	if err == nil {
		t.Fatalf("DomainExist failed")
	}
}

func TestNewDomain(t *testing.T) {
	err := NewDomain(context.Background(), "")
	if err == nil {
		t.Fatalf("NewDomain failed")
	}
}

func TestProjectExist(t *testing.T) {
	_, err := ProjectExist(util.SetContext(context.Background(), CTX_CACHEONLY, "1"), "", "")
	if err != nil {
		t.Fatalf("DomainExist WithCacheOnly failed")
	}

	_, err = ProjectExist(context.Background(), "", "")
	if err == nil {
		t.Fatalf("DomainExist failed")
	}
}

func TestNewProject(t *testing.T) {
	err := NewProject(context.Background(), "", "")
	if err == nil {
		t.Fatalf("NewProject failed")
	}
}

func TestNewDomainProject(t *testing.T) {
	err := NewDomainProject(context.Background(), "", "")
	if err == nil {
		t.Fatalf("NewDomainProject failed")
	}
}

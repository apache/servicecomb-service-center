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

package util_test

import (
	"context"
	"net/http"
	"testing"

	. "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/go-chassis/cari/discovery"
)

func TestAllowAcrossApp(t *testing.T) {
	err := AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
	}, &discovery.MicroService{
		AppId: "a",
	})
	if err != nil {
		t.Fatalf("AllowAcrossApp with the same appId and no property failed")
	}

	err = AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
	}, &discovery.MicroService{
		AppId: "c",
	})
	if err == nil {
		t.Fatalf("AllowAcrossApp with the diff appId and no property failed")
	}

	err = AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
		Properties: map[string]string{
			discovery.PropAllowCrossApp: "true",
		},
	}, &discovery.MicroService{
		AppId: "a",
	})
	if err != nil {
		t.Fatalf("AllowAcrossApp with the same appId and allow property failed")
	}

	err = AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
		Properties: map[string]string{
			discovery.PropAllowCrossApp: "true",
		},
	}, &discovery.MicroService{
		AppId: "b",
	})
	if err != nil {
		t.Fatalf("AllowAcrossApp with the diff appId and allow property failed")
	}

	err = AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
		Properties: map[string]string{
			discovery.PropAllowCrossApp: "false",
		},
	}, &discovery.MicroService{
		AppId: "b",
	})
	if err == nil {
		t.Fatalf("AllowAcrossApp with the diff appId and deny property failed")
	}

	err = AllowAcrossDimension(context.Background(), &discovery.MicroService{
		AppId: "a",
		Properties: map[string]string{
			discovery.PropAllowCrossApp: "",
		},
	}, &discovery.MicroService{
		AppId: "b",
	})
	if err == nil {
		t.Fatalf("AllowAcrossApp with the diff appId and empty property failed")
	}
}

func TestGetConsumer(t *testing.T) {
	_, err := GetConsumerIds(context.Background(), "",
		&discovery.MicroService{
			ServiceId: "a",
		})
	if err != nil {
		t.Fatalf("GetConsumerIds WithCacheOnly failed")
	}
}

func TestGetProvider(t *testing.T) {
	_, err := GetProviderIds(context.Background(), "",
		&discovery.MicroService{
			ServiceId: "a",
		})
	if err != nil {
		t.Fatalf("GetProviderIds WithCacheOnly failed")
	}
}

func TestAccessible(t *testing.T) {
	err := Accessible(context.Background(), "xxx", "")
	if err.StatusCode() != http.StatusBadRequest {
		t.Fatalf("Accessible invalid failed")
	}
}

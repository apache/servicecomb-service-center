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

package test

import (
	"time"

	"github.com/go-chassis/cari/db"
	"github.com/go-chassis/cari/db/config"
	"github.com/go-chassis/go-archaius"

	_ "github.com/apache/servicecomb-service-center/eventbase/bootstrap"
	_ "github.com/go-chassis/cari/db/bootstrap"
)

var (
	Etcd             = "etcd"
	EtcdURI          = "http://127.0.0.1:2379"
	Mongo            = "mongo"
	MongoURI         = "mongodb://127.0.0.1:27017"
	DefaultTestDB    = "etcd"
	DefaultTestDBURI = "http://127.0.0.1:2379"
)

var DBKind string

func init() {
	cfg := config.Config{}
	err := archaius.Init(archaius.WithMemorySource(), archaius.WithENVSource())
	if err != nil {
		panic(err)
	}
	mode, ok := archaius.Get("TEST_DB_MODE").(string)
	if ok {
		DefaultTestDB = mode
	}
	uri, ok := archaius.Get("TEST_DB_URI").(string)
	if ok {
		DefaultTestDBURI = uri
	}
	DBKind = DefaultTestDB
	cfg.Kind = DefaultTestDB
	cfg.URI = DefaultTestDBURI
	cfg.Timeout = 10 * time.Second
	err = db.Init(&cfg)
	if err != nil {
		panic(err)
	}
}

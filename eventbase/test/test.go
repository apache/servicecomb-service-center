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
	"github.com/go-chassis/go-archaius"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/dlock"

	_ "github.com/apache/servicecomb-service-center/eventbase/bootstrap"
)

var (
	Etcd             = "etcd"
	EtcdURI          = "http://127.0.0.1:2379"
	Mongo            = "mongo"
	MongoURI         = "mongodb://127.0.0.1:27017"
	DefaultTestDB    = "etcd"
	DefaultTestDBURI = "http://127.0.0.1:2379"
)

var DbCfg = db.Config{}

func init() {
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
	DbCfg.Kind = DefaultTestDB
	DbCfg.URI = DefaultTestDBURI
	DbCfg.Timeout = 10 * time.Second
	err = datasource.Init(DbCfg)
	if err != nil {
		panic(err)
	}
	_ = dlock.Init(dlock.Options{
		Kind: DbCfg.Kind,
	})
}

func IsETCD() bool {
	t := archaius.Get("TEST_DB_MODE")
	if t == nil {
		t = "etcd"
	}
	return t == "etcd"
}

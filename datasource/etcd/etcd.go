// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"errors"
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

// TODO: define error with names here

var ErrNotUnique = errors.New("kv result is not unique")

func init() {
	// TODO: set logger
	// TODO: register storage plugin to plugin manager
}

type DataSource struct {
	// schemaEditable determines whether schema modification is allowed for
	SchemaEditable bool
	ttlFromEnv     int64
}

func NewDataSource(opts datasource.Options) *DataSource {
	// TODO: construct a reasonable DataSource instance
	log.Warnf("dependency data source enable etcd mode")

	inst := &DataSource{
		SchemaEditable: opts.SchemaEditable,
		ttlFromEnv:     opts.TTL,
	}
	// TODO: deal with exception
	if err := inst.initialize(); err != nil {
		return inst
	}
	return inst
}

func (ds *DataSource) initialize() error {
	// TODO: init dependency members
	return nil
}

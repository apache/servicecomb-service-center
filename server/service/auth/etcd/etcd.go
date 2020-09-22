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
	"github.com/apache/servicecomb-service-center/pkg/log"
)

// TODO: define error with names here

func init() {
	// TODO: set logger
	// TODO: register storage plugin to plugin manager
}

type DataSource struct{}

func NewDataSource() *DataSource {
	// TODO: construct a reasonable DataSource instance
	log.Warnf("auth data source enable etcd mode")

	inst := &DataSource{}
	// TODO: deal with exception
	if err := inst.initialize(); err != nil {
		return inst
	}
	return inst
}

func (ds *DataSource) initialize() error {
	// TODO: init DataSource members
	return nil
}

func (ds *DataSource) AddAccount() {
	panic("implement me")
}

func (ds *DataSource) GetAccount() {
	panic("implement me")
}

func (ds *DataSource) UpdateAccount() {
	panic("implement me")
}

func (ds *DataSource) DeleteAccount() {
	panic("implement me")
}

func (ds *DataSource) AddDomain() {
	panic("implement me")
}

func (ds *DataSource) GetDomain() {
	panic("implement me")
}

func (ds *DataSource) UpdateDomain() {
	panic("implement me")
}

func (ds *DataSource) DeleteDomain() {
	panic("implement me")
}

func (ds *DataSource) AddDomainProject() {
	panic("implement me")
}

func (ds *DataSource) GetDomainProject() {
	panic("implement me")
}

func (ds *DataSource) UpdateDomainProject() {
	panic("implement me")
}

func (ds *DataSource) DeleteDomainProject() {
	panic("implement me")
}

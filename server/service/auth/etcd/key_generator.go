/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package etcd

import "github.com/apache/servicecomb-service-center/pkg/util"

const (
	RegistryRootKey    = "cse-sr"
	RegistryProjectKey = "projects"
	SPLIT              = "/"
	RegistryDomainKey  = "domains"
)

func GetRootKey() string {
	return SPLIT + RegistryRootKey
}

func (ds *DataSource) GenerateAccountKey(name string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		"accounts",
		name,
	}, SPLIT)
}

func GenerateETCDAccountKey(name string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		"accounts",
		name,
	}, SPLIT)
}

func GenerateETCDProjectKey(domain, project string) string {
	return util.StringJoin([]string{
		GetProjectRootKey(domain),
		project,
	}, SPLIT)
}

func GetProjectRootKey(domain string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryProjectKey,
		domain,
	}, SPLIT)
}

func GenerateETCDDomainKey(domain string) string {
	return util.StringJoin([]string{
		GetDomainRootKey(),
		domain,
	}, SPLIT)
}

func GetDomainRootKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryDomainKey,
	}, SPLIT)
}

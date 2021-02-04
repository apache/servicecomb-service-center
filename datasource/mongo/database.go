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

package mongo

import (
	"time"

	pb "github.com/go-chassis/cari/discovery"
)

const (
	CollectionAccount  = "account"
	CollectionService  = "service"
	CollectionSchema   = "schema"
	CollectionRule     = "rule"
	CollectionInstance = "instance"
	CollectionDep      = "dependency"
	CollectionRole     = "role"
)

const (
	DepsQueueUUID = "0"
)

const (
	ColumnDomain              = "domain"
	ColumnProject             = "project"
	ColumnTag                 = "tags"
	ColumnSchemaID            = "schemaid"
	ColumnServiceID           = "serviceid"
	ColumnRuleID              = "ruleid"
	ColumnService             = "service"
	ColumnProperty            = "properties"
	ColumnModTime             = "modtimestamp"
	ColumnEnv                 = "env"
	ColumnAppID               = "app"
	ColumnServiceName         = "servicename"
	ColumnAlias               = "alias"
	ColumnVersion             = "version"
	ColumnSchemas             = "schemas"
	ColumnAttribute           = "attribute"
	ColumnPattern             = "pattern"
	ColumnDescription         = "description"
	ColumnRuleType            = "ruletype"
	ColumnSchema              = "schema"
	ColumnSchemaSummary       = "schemasummary"
	ColumnDep                 = "dep"
	ColumnDependency          = "dependency"
	ColumnRule                = "rule"
	ColumnInstance            = "instance"
	ColumnInstanceID          = "instanceid"
	ColumnConsumerID          = "consumerid"
	ColumnMongoID             = "_id"
	ColumnTenant              = "tenant"
	ColumnServiceType         = "type"
	ColumnServiceKey          = "servicekey"
	ColumnConsumer            = "consumer"
	ColumnID                  = "id"
	ColumnAccountName         = "name"
	ColumnRoleName            = "name"
	ColumnPerms               = "perms"
	ColumnPassword            = "password"
	ColumnRole                = "role"
	ColumnTokenExpirationTime = "tokenexpirationtime"
	ColumnCurrentPassword     = "currentpassword"
	ColumnStatus              = "status"
	ColumnRefreshTime         = "refreshtime"
)

type Service struct {
	Domain  string
	Project string
	Tags    map[string]string
	Service *pb.MicroService
}

type Schema struct {
	Domain        string
	Project       string
	ServiceID     string
	SchemaID      string
	Schema        string
	SchemaSummary string
}

type Rule struct {
	Domain    string
	Project   string
	ServiceID string
	Rule      *pb.ServiceRule
}

type Instance struct {
	Domain      string
	Project     string
	RefreshTime time.Time
	Instance    *pb.MicroServiceInstance
}

type ConsumerDep struct {
	Domain      string
	Project     string
	ConsumerID  string
	UUID        string
	ConsumerDep *pb.ConsumerDependency
}

type DependencyRule struct {
	Type       string
	Domain     string
	Project    string
	ServiceKey *pb.MicroServiceKey
	Dep        *pb.MicroServiceDependency
}

type DelDepCacheKey struct {
	key  *pb.MicroServiceKey
	Type string
}

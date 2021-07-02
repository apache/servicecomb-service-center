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

package model

import (
	"time"

	pb "github.com/go-chassis/cari/discovery"
)

const (
	CollectionAccount     = "account"
	CollectionAccountLock = "account_lock"
	CollectionService     = "service"
	CollectionSchema      = "schema"
	CollectionRule        = "rule"
	CollectionInstance    = "instance"
	CollectionDep         = "dependency"
	CollectionRole        = "role"
	CollectionDomain      = "domain"
	CollectionProject     = "project"
)

const (
	ColumnDomain               = "domain"
	ColumnProject              = "project"
	ColumnTag                  = "tags"
	ColumnSchemaID             = "schema_id"
	ColumnServiceID            = "service_id"
	ColumnRuleID               = "rule_id"
	ColumnService              = "service"
	ColumnProperty             = "properties"
	ColumnModTime              = "mod_timestamp"
	ColumnEnv                  = "env"
	ColumnAppID                = "app"
	ColumnServiceName          = "service_name"
	ColumnAlias                = "alias"
	ColumnVersion              = "version"
	ColumnSchemas              = "schemas"
	ColumnAttribute            = "attribute"
	ColumnPattern              = "pattern"
	ColumnDescription          = "description"
	ColumnRuleType             = "rule_type"
	ColumnSchema               = "schema"
	ColumnSchemaSummary        = "schema_summary"
	ColumnDep                  = "dep"
	ColumnDependency           = "dependency"
	ColumnRule                 = "rule"
	ColumnInstance             = "instance"
	ColumnInstanceID           = "instance_id"
	ColumnTenant               = "tenant"
	ColumnServiceType          = "type"
	ColumnServiceKey           = "service_key"
	ColumnID                   = "id"
	ColumnAccountName          = "name"
	ColumnRoleName             = "name"
	ColumnPerms                = "perms"
	ColumnAccountUpdateTime    = "updatetime"
	ColumnRoleUpdateTime       = "updatetime"
	ColumnPassword             = "password"
	ColumnRoles                = "roles"
	ColumnTokenExpirationTime  = "token_expiration_time"
	ColumnCurrentPassword      = "current_password"
	ColumnStatus               = "status"
	ColumnRefreshTime          = "refresh_time"
	ColumnAccountLockKey       = "key"
	ColumnAccountLockStatus    = "status"
	ColumnAccountLockReleaseAt = "release_at"
)

type Service struct {
	Domain  string            `json:"domain,omitempty"`
	Project string            `json:"project,omitempty"`
	Tags    map[string]string `json:"tags,omitempty"`
	Service *pb.MicroService  `json:"service,omitempty"`
}

type Schema struct {
	Domain        string `json:"domain,omitempty"`
	Project       string `json:"project,omitempty"`
	ServiceID     string `json:"serviceID,omitempty" bson:"service_id"`
	SchemaID      string `json:"schemaID,omitempty" bson:"schema_id"`
	Schema        string `json:"schema,omitempty"`
	SchemaSummary string `json:"schemaSummary,omitempty" bson:"schema_summary"`
}

type Rule struct {
	Domain    string          `json:"domain,omitempty"`
	Project   string          `json:"project,omitempty"`
	ServiceID string          `json:"serviceID,omitempty" bson:"service_id"`
	Rule      *pb.ServiceRule `json:"rule,omitempty"`
}

type Instance struct {
	Domain      string                   `json:"domain,omitempty"`
	Project     string                   `json:"project,omitempty"`
	RefreshTime time.Time                `json:"refreshTime,omitempty" bson:"refresh_time"`
	Instance    *pb.MicroServiceInstance `json:"instance,omitempty"`
}

type ConsumerDep struct {
	Domain      string                 `json:"domain,omitempty"`
	Project     string                 `json:"project,omitempty"`
	ConsumerID  string                 `json:"consumerID,omitempty" bson:"consumer_id"`
	UUID        string                 `json:"uuID,omitempty" bson:"uu_id"`
	ConsumerDep *pb.ConsumerDependency `json:"consumerDep,omitempty" bson:"consumer_dep"`
}

type DependencyRule struct {
	Type       string                     `json:"type,omitempty"`
	Domain     string                     `json:"domain,omitempty"`
	Project    string                     `json:"project,omitempty"`
	ServiceKey *pb.MicroServiceKey        `json:"serviceKey,omitempty" bson:"service_key"`
	Dep        *pb.MicroServiceDependency `json:"dep,omitempty"`
}

type DelDepCacheKey struct {
	Key  *pb.MicroServiceKey
	Type string
}

type Domain struct {
	Domain string `json:"domain,omitempty"`
}

type Project struct {
	Domain  string `json:"domain,omitempty"`
	Project string `json:"project,omitempty"`
}

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
	pb "github.com/go-chassis/cari/discovery"
	"time"
)

const (
	DuplicateKey               = 11000
	AccountName                = "name"
	AccountID                  = "id"
	AccountPassword            = "password"
	AccountRole                = "role"
	AccountTokenExpirationTime = "tokenexpirationtime"
	AccountCurrentPassword     = "currentpassword"
	AccountStatus              = "status"
	InstanceID                 = "instanceinfo.instanceid"
	ServiceID                  = "instanceinfo.serviceid"
	RefreshTime                = "refreshtime"
)

const (
	CollectionAccount  = "account"
	CollectionService  = "service"
	CollectionSchema   = "schema"
	CollectionRule     = "rule"
	CollectionInstance = "instance"
)

const (
	ErrorDuplicateKey = 11000
)

const (
	Domain             = "domain"
	Project            = "project"
	ServiceTag         = "tags"
	SchemaID           = "schemaid"
	RuleServiceID      = "serviceid"
	RuleRuleID         = "ruleinfo.ruleid"
	SchemaServiceID    = "serviceid"
	ServiceServiceID   = "serviceinfo.serviceid"
	ServiceProperty    = "serviceinfo.properties"
	ServiceModTime     = "serviceinfo.modtimestamp"
	ServiceEnv         = "serviceinfo.environment"
	ServiceAppID       = "serviceinfo.appid"
	ServiceServiceName = "serviceinfo.servicename"
	ServiceAlias       = "serviceinfo.alias"
	ServiceVersion     = "serviceinfo.version"
	ServiceSchemas     = "serviceinfo.schemas"
	RuleAttribute      = "ruleinfo.attribute"
	RulePattern        = "ruleinfo.pattern"
	RuleModTime        = "ruleinfo.modtimestamp"
	RuleDescription    = "ruleinfo.description"
	RuleRuletype       = "ruleinfo.ruletype"
	SchemaInfo         = "schemainfo"
	SchemaSummary      = "schemasummary"
)

type Service struct {
	Domain      string
	Project     string
	Tags        map[string]string
	ServiceInfo *pb.MicroService
}

type Schema struct {
	Domain        string
	Project       string
	ServiceID     string
	SchemaID      string
	SchemaInfo    string
	SchemaSummary string
}

type Rule struct {
	Domain    string
	Project   string
	ServiceID string
	RuleInfo  *pb.ServiceRule
}

type Instance struct {
	Domain       string
	Project      string
	RefreshTime  time.Time
	InstanceInfo *pb.MicroServiceInstance
}

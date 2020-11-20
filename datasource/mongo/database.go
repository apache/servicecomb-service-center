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
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
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
	CollectionInstance         = "instance"
	InstanceID                 = "instanceinfo.instanceid"
	ServiceID                  = "instanceinfo.serviceid"
	RefreshTime                = "refreshtime"
)

const (
	CollectionAccount = "account"
	CollectionService = "service"
	CollectionSchema  = "schema"
	CollectionRule    = "rule"
)

const (
	ErrorDuplicateKey = 11000
)

const (
	Domain           = "domain"
	Project          = "project"
	ServiceTag       = "tags"
	SchemaID         = "schemaid"
	RuleServiceID    = "serviceid"
	RuleRuleID       = "rule.ruleid"
	SchemaServiceID  = "serviceid"
	ServiceServiceID = "service.serviceid"
	ServiceProperty  = "service.properties"
	ServiceModTime   = "service.modtimestamp"
	ServiceSchemas   = "service.schemas"
	RuleAttribute    = "rule.attribute"
	RulePattern      = "rule.pattern"
	RuleModTime      = "rule.modtimestamp"
	RuleDescription  = "rule.description"
	RuleRuletype     = "rule.ruletype"
	Schema           = "schema"
	SchemaSummary    = "schemasummary"
)

type MgService struct {
	Domain  string
	Project string
	Tags    map[string]string
	Service *pb.MicroService
}

type MgSchema struct {
	Domain        string
	Project       string
	ServiceID     string
	SchemaID      string
	Schema        string
	SchemaSummary string
}

type MgRule struct {
	Domain    string
	Project   string
	ServiceID string
	Rule      *pb.ServiceRule
}

type Instance struct {
	Domain       string
	Project      string
	RefreshTime  time.Time
	InstanceInfo *pb.MicroServiceInstance
}

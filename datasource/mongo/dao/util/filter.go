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

package util

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"github.com/go-chassis/cari/rbac"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"github.com/apache/servicecomb-service-center/pkg/util"
)

type Option func(filter bson.M)

func Domain(domain string) Option {
	return func(filter bson.M) {
		filter[dao.ColumnDomain] = domain
	}
}

func Project(project string) Option {
	return func(filter bson.M) {
		filter[dao.ColumnProject] = project
	}
}







func Status(status string) Option {
	return func(filter bson.M) {
		filter[dao.ColumnStatus] = status
	}
}

func ID(id string) Option {
	return func(filter bson.M) {
		filter[dao.ColumnID] = id
	}
}

func RoleName(name string) Option {
	return func(filter bson.M) {
		filter[dao.ColumnRoleName] = name
	}
}

func Perms(perms []*rbac.Permission) Option {
	return func(filter bson.M) {
		filter[dao.ColumnPerms] = perms
	}
}

func In(data interface{}) Option {
	return func(filter bson.M) {
		filter["$in"] = data
	}
}

func Set(data interface{}) Option {
	return func(filter bson.M) {
		filter["$set"] = data
	}
}

func NewFilter(options ...func(filter bson.M)) bson.M {
	filter := bson.M{}
	for _, option := range options {
		option(filter)
	}
	return filter
}

func NewDomainProjectFilter(domain string, project string, options ...func(filter bson.M)) bson.M {
	filter := bson.M{
		dao.ColumnDomain:  domain,
		dao.ColumnProject: project,
	}
	for _, option := range options {
		option(filter)
	}
	return filter
}

func NewBasicFilter(ctx context.Context, options ...func(filter bson.M)) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	filter := bson.M{
		dao.ColumnDomain:  domain,
		dao.ColumnProject: project,
	}
	for _, option := range options {
		option(filter)
	}
	return filter
}

func InstanceServiceID(serviceID interface{}) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnInstance, dao.ColumnServiceID})] = serviceID
	}
}

func InstanceInstanceID(instanceID string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnInstance, dao.ColumnInstanceID})] = instanceID
	}
}

func ServiceServiceID(serviceID string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnService, dao.ColumnServiceID})] = serviceID
	}
}

func ServiceEnv(env string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnService, dao.ColumnEnv})] = env
	}
}

func ServiceAppID(appID string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnService, dao.ColumnAppID})] = appID
	}
}

func ServiceModTime(modTime string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnService, dao.ColumnModTime})] = modTime
	}
}

func ServiceProperty(property map[string]string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnService, dao.ColumnProperty})] = property
	}
}

func ServiceServiceName(serviceName string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnService, dao.ColumnServiceName})] = serviceName
	}
}

func ServiceID(serviceID string) Option {
	return func(filter bson.M) {
		filter[dao.ColumnServiceID] = serviceID
	}
}

func ServiceAlias(alias string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnService, dao.ColumnAlias})] = alias
	}
}

func ServiceSchemas(schemas []string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnService, dao.ColumnSchemas})] = schemas
	}
}

func ServiceVersion(version interface{}) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnService, dao.ColumnVersion})] = version
	}
}

func ServiceType(serviceType string) Option {
	return func(filter bson.M) {
		filter[dao.ColumnServiceType] = serviceType
	}
}

func ServiceKeyTenant(tenant string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnServiceKey, dao.ColumnTenant})] = tenant
	}
}

func ServiceKeyAppID(appID string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnServiceKey, dao.ColumnAppID})] = appID
	}
}

func ServiceKeyServiceName(serviceName string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnServiceKey, dao.ColumnServiceName})] = serviceName
	}
}

func ServiceKeyServiceEnv(env string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnServiceKey, dao.ColumnEnv})] = env
	}
}

func ServiceKeyServiceVersion(version string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnServiceKey, dao.ColumnVersion})] = version
	}
}

func Schema(schema string) Option {
	return func(filter bson.M) {
		filter[dao.ColumnSchema] = schema
	}
}

func SchemaID(schemaID string) Option {
	return func(filter bson.M) {
		filter[dao.ColumnSchemaID] = schemaID
	}
}

func RuleAttribute(attribute string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnRule, dao.ColumnAttribute})] = attribute
	}
}

func RuleRuleID(ruleID string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnRule, dao.ColumnRuleID})] = ruleID
	}
}

func RuleRuleType(ruleType string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnRule, dao.ColumnRuleType})] = ruleType
	}
}

func RulePattern(pattern string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnRule, dao.ColumnPattern})] = pattern
	}
}

func RuleDescription(description string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnRule, dao.ColumnDescription})] = description
	}
}

func RuleModTime(modTime string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnRule, dao.ColumnModTime})] = modTime
	}
}

func SchemaSummary(schemaSummary string) Option {
	return func(filter bson.M) {
		filter[dao.ColumnSchemaSummary] = schemaSummary
	}
}

func Tags(tags map[string]string) Option {
	return func(filter bson.M) {
		filter[dao.ColumnTag] = tags
	}
}

func InstanceModTime(modTime string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnService, dao.ColumnModTime})] = modTime
	}
}

func InstanceStatus(status string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnInstance, dao.ColumnStatus})] = status
	}
}

func InstanceProperties(properties map[string]string) Option {
	return func(filter bson.M) {
		filter[util2.ConnectWithDot([]string{dao.ColumnInstance, dao.ColumnProperty})] = properties
	}
}

func BuildIndexDoc(keys ...string) mongo.IndexModel {
	keysDoc := bsonx.Doc{}
	for _, key := range keys {
		keysDoc = keysDoc.Append(key, bsonx.Int32(1))
	}
	index := mongo.IndexModel{
		Keys: keysDoc,
	}
	return index
}

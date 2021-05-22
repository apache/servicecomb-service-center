/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except request compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to request writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package heartbeat

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/util"
)

var (
	ErrPluginNameNil    = errors.New("plugin implement name is nil")
	ErrPluginNotSupport = errors.New("plugin implement not supported [#{opts.Kind}]")
)

func NewServiceIDInstanceIDFilter(serviceID string, instanceID string) bson.D {
	filter := bson.D{
		{util.ConnectWithDot([]string{dao.ColumnInstance, dao.ColumnServiceID}), serviceID},
		{util.ConnectWithDot([]string{dao.ColumnInstance, dao.ColumnInstanceID}), instanceID},
	}
	return filter
}

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

package event

import (
	"context"
	"encoding/json"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
)

func Publish(ctx context.Context, action string, resourceType string, resource interface{}) {
	eventID, err := v1sync.NewEventID()
	if err != nil {
		log.Error("fail to create eventID", err)
		return
	}
	resourceValue, err := json.Marshal(resource)
	if err != nil {
		log.Error("fail to marshal the resource", err)
		return
	}

	e := &v1sync.Event{
		Id: eventID,
		Opts: map[string]string{
			string(util.CtxDomain):  util.ParseDomain(ctx),
			string(util.CtxProject): util.ParseProject(ctx),
		},
		Subject:   resourceType,
		Action:    action,
		Value:     resourceValue,
		Timestamp: v1sync.Timestamp(),
	}
	Send(&Event{
		Event: e,
	})
}

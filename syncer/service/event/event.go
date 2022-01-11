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
	"encoding/json"
	"fmt"

	guuid "github.com/gofrs/uuid"

	v1 "github.com/apache/servicecomb-service-center/api/sync/v1"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func Publish(action string, resourceType string, resource interface{}) {
	eventID, err := guuid.NewV4()
	if err != nil {
		log.Error("fail to create eventID", err)
		return
	}
	resourceValue, err := json.Marshal(resource)
	if err != nil {
		log.Error("fail to marshal the resource", err)
		return
	}
	event := v1.Event{
		Id:      eventID.String(),
		Action:  action,
		Subject: resourceType,
		Value:   resourceValue,
	}
	log.Info(fmt.Sprintf("success to send event %s", event.Subject))
	// TODO to send event
}

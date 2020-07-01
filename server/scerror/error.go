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
package scerror

import (
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

var errors = map[int32]string{
	ErrInvalidParams:           "Invalid parameter(s)",
	ErrUnhealthy:               "Server is Unhealthy",
	ErrServiceAlreadyExists:    "Micro-service already exists",
	ErrServiceNotExists:        "Micro-service does not exist",
	ErrServiceVersionNotExists: "Micro-service version does not exist",
	ErrDeployedInstance:        "Micro-service has deployed instance(s)",
	ErrDependedOnConsumer:      "Consumer(s) depends on this micro-service",
	ErrUndefinedSchemaId:       "Undefined schema id",
	ErrModifySchemaNotAllow:    "Not allowed to modify schema",
	ErrSchemaNotExists:         "Schema does not exist",
	ErrInstanceNotExists:       "Instance does not exist",
	ErrPermissionDeny:          "Access micro-service refused",
	ErrTagNotExists:            "Tag does not exist",
	ErrRuleAlreadyExists:       "Rule already exist",
	ErrBlackAndWhiteRule:       "Can not have both 'BLACK' and 'WHITE'",
	ErrModifyRuleNotAllow:      "Not allowed to modify the type of the rule",
	ErrRuleNotExists:           "Rule does not exist",
	ErrNotEnoughQuota:          "Not enough quota",
	ErrUnauthorized:            "Request unauthorized",
	ErrInternal:                "Internal server error",
	ErrUnavailableBackend:      "Registry service is unavailable",
	ErrUnavailableQuota:        "Quota service is unavailable",
	ErrEndpointAlreadyExists:   "Endpoint is already belong to other service",
	ErrForbidden:               "Forbidden",
}

const (
	ErrInvalidParams           int32 = 400001
	ErrUnhealthy               int32 = 400002
	ErrServiceAlreadyExists    int32 = 400010
	ErrServiceNotExists        int32 = 400012
	ErrDeployedInstance        int32 = 400013
	ErrUndefinedSchemaId       int32 = 400014
	ErrModifySchemaNotAllow    int32 = 400015
	ErrSchemaNotExists         int32 = 400016
	ErrInstanceNotExists       int32 = 400017
	ErrTagNotExists            int32 = 400018
	ErrRuleAlreadyExists       int32 = 400019
	ErrBlackAndWhiteRule       int32 = 400020
	ErrModifyRuleNotAllow      int32 = 400021
	ErrRuleNotExists           int32 = 400022
	ErrDependedOnConsumer      int32 = 400023
	ErrPermissionDeny          int32 = 400024
	ErrEndpointAlreadyExists   int32 = 400025
	ErrServiceVersionNotExists int32 = 400026
	ErrNotEnoughQuota          int32 = 400100
	ErrUnauthorized            int32 = 401002
	ErrForbidden               int32 = 403001
	ErrInternal                int32 = 500003
	ErrUnavailableBackend      int32 = 500011
	ErrUnavailableQuota        int32 = 500101
)

type Error struct {
	Code    int32  `json:"errorCode,string"`
	Message string `json:"errorMessage"`
	Detail  string `json:"detail,omitempty"`
}

func (e *Error) Error() string {
	if len(e.Detail) == 0 {
		return e.Message
	}
	return e.Message + "(" + e.Detail + ")"
}

func (e *Error) Marshal() []byte {
	bs, _ := json.Marshal(e)
	return bs
}

func (e *Error) StatusCode() int {
	return int(e.Code / 1000)
}

func (e *Error) InternalError() bool {
	return e.Code >= 500000
}

func NewError(code int32, detail string) *Error {
	return &Error{
		Code:    code,
		Message: errors[code],
		Detail:  detail,
	}
}

func NewErrorf(code int32, format string, args ...interface{}) *Error {
	return NewError(code, fmt.Sprintf(format, args...))
}

func RegisterErrors(errs map[int32]string) {
	for err, msg := range errs {
		if err < 400000 || err >= 600000 {
			log.Warnf("error code[%v] should be between 4xx and 5xx", err)
			continue
		}
		errors[err] = msg
	}
}

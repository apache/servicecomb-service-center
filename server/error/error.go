//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package error

import (
	"encoding/json"
	"fmt"
	"github.com/ServiceComb/service-center/pkg/util"
	"net/http"
)

var errors = map[int32]string{
	ErrInvalidParams: "Invalid parameter(s)",

	ErrServiceAlreadyExists: "Micro-service already exists",
	ErrServiceNotExists:     "Micro-service does not exist",
	ErrDeployedInstance:     "Micro-service has deployed instance(s)",
	ErrDependedOnConsumer:   "Consumer(s) depends on this micro-service",

	ErrUndefinedSchemaId:    "Undefined schema id",
	ErrModifySchemaNotAllow: "Not allowed to modify schema",
	ErrSchemaNotExists:      "Schema does not exist",

	ErrInstanceNotExists: "Instance does not exist",
	ErrPermissionDeny:    "Access micro-service refused",

	ErrTagNotExists: "Tag does not exist",

	ErrRuleAlreadyExists:  "Rule already exist",
	ErrBlackAndWhiteRule:  "Can not have both 'BLACK' and 'WHITE'",
	ErrModifyRuleNotAllow: "Not allowed to modify the type of the rule",
	ErrRuleNotExists:      "Rule does not exist",

	ErrNotEnoughQuota: "Not enough quota",

	ErrUnauthorized: "Request unauthorized",

	ErrInternal:           "Internal server error",
	ErrUnavailableBackend: "Registry service is unavailable",
	ErrUnavailableQuota:   "Quota service is unavailable",

	ErrEnpointsMoreBelongToOtherService: "endpoint more belong to other service",
}

const (
	ErrInvalidParams int32 = 400001
	ErrUnauthorized  int32 = 401002
	ErrInternal      int32 = 500003

	ErrServiceAlreadyExists int32 = 400010
	ErrUnavailableBackend   int32 = 500011

	ErrServiceNotExists int32 = 400012

	ErrDeployedInstance int32 = 400013

	ErrUndefinedSchemaId    int32 = 400014
	ErrModifySchemaNotAllow int32 = 400015
	ErrSchemaNotExists      int32 = 400016

	ErrInstanceNotExists int32 = 400017

	ErrTagNotExists int32 = 400018

	ErrRuleAlreadyExists  int32 = 400019
	ErrBlackAndWhiteRule  int32 = 400020
	ErrModifyRuleNotAllow int32 = 400021
	ErrRuleNotExists      int32 = 400022

	ErrDependedOnConsumer int32 = 400023

	ErrPermissionDeny int32 = 400024

	ErrEnpointsMoreBelongToOtherService int32 = 400025

	ErrNotEnoughQuota   int32 = 400100
	ErrUnavailableQuota int32 = 500101
)

type Error struct {
	Code    int32  `json:"errorCode,string"`
	Message string `json:"errorMessage"`
	Detail  string `json:"detail,omitempty"`
}

func (e Error) Error() string {
	if len(e.Detail) == 0 {
		return e.Message
	}
	return e.Message + "(" + e.Detail + ")"
}

func (e Error) toJson() string {
	bs, _ := json.Marshal(e)
	return util.BytesToStringWithNoCopy(bs)
}

func (e Error) StatusCode() int {
	if e.Code >= 500000 {
		return http.StatusInternalServerError
	}
	return http.StatusBadRequest
}

func (e Error) HttpWrite(w http.ResponseWriter) {
	status := e.StatusCode()
	w.Header().Add("X-Response-Status", fmt.Sprint(status))
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	fmt.Fprintln(w, e.toJson())
}

func NewError(code int32, detail string) *Error {
	return &Error{
		Code:    code,
		Message: errors[code],
		Detail:  detail,
	}
}

func RegisterErrors(errs map[int32]string) {
	for err, msg := range errs {
		if err < 400000 || err >= 600000 {
			// should be between 4xx and 5xx
			continue
		}
		errors[err] = msg
	}
}

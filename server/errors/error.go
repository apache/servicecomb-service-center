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
package errors

var errors = map[int]string{
	ErrInvalidParams: "Invalid parameter(s)",

	ErrServiceAlreadyExists: "Micro-service already exists",
	ErrServiceNotExists:     "Micro-service does not exist",
	ErrDeployedInstance:     "Micro-service has deployed instance(s)",

	ErrUndefinedSchemaId:    "Undefined schema id",
	ErrModifySchemaNotAllow: "Not allowed to modify schema",
	ErrSchemaNotExists:      "Schema does not exist",

	ErrInstanceNotExists: "Instance does not exist",

	ErrTagNotExists: "Tag does not exist",

	ErrRuleAlreadyExists:  "Rule already exist",
	ErrBlackAndWhiteRule:  "Can not have both 'BLACK' and 'WHITE'",
	ErrModifyRuleNotAllow: "Not allowed to modify the type of the rule",
	ErrRuleNotExists:      "Rule does not exist",

	ErrNotEnoughQuota: "Not enough quota",

	ErrUnauthorized: "Request unauthorized",

	ErrInternalException:  "Internal server error",
	ErrUnavailableBackend: "Registry service is unavailable",
	ErrUnavailableQuota:   "Quota service is unavailable",
}

const (
	ErrInvalidParams     = 400001
	ErrUnauthorized      = 401002
	ErrInternalException = 500003

	ErrServiceAlreadyExists = 400010
	ErrUnavailableBackend   = 500011

	ErrServiceNotExists = 400012

	ErrDeployedInstance = 400013

	ErrUndefinedSchemaId    = 400014
	ErrModifySchemaNotAllow = 400015
	ErrSchemaNotExists      = 400016

	ErrInstanceNotExists = 400017

	ErrTagNotExists = 400018

	ErrRuleAlreadyExists  = 400019
	ErrBlackAndWhiteRule  = 400020
	ErrModifyRuleNotAllow = 400021
	ErrRuleNotExists      = 400022

	ErrNotEnoughQuota   = 400100
	ErrUnavailableQuota = 500101
)

type Error struct {
	Code    int    `json:"errCode,string"`
	Message string `json:"errMsg"`
	Detail  string `json:"detail,omitempty"`
}

func (e Error) Error() string {
	if len(e.Detail) == 0 {
		return e.Message
	}
	return e.Message + "(" + e.Detail + ")"
}

func NewError(code int, detail string) *Error {
	return &Error{
		Code:    code,
		Message: errors[code],
		Detail:  detail,
	}
}

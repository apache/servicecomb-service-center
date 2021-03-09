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

package util

import "errors"

var (
	ErrInvalidConsumer                      = errors.New("Invalid consumer")
	ErrUpdateNodata                         = errors.New("UpdateAccount: no data to update")
	ErrServiceFileLost                      = errors.New("service center service file lost")
	ErrInvalidDomainProject                 = errors.New("invalid domainProject")
	ErrNotAllowDeleteSC                     = errors.New("not allow to delete service center")
	ErrDeleteSchemaFailed                   = errors.New("delete schema failed")
	ErrInvalidParamBatchGetInstancesRequest = errors.New("invalid param BatchGetInstancesRequest")
)

func NewError(errInfo string, errMsg string) error {
	return errors.New(errInfo + errMsg)
}

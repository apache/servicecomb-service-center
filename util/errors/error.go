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

import (
	original "errors"
	"fmt"
)

func New(message string) error {
	return original.New(message)
}

func NewWithFmt(message string, args ...interface{}) error {
	return original.New(fmt.Sprintf(message, args...))
}

func NewWithError(message string, err error) error {
	return NewWithFmt("%s: %s", message, err.Error())
}

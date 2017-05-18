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

import "fmt"

type HttpError interface {
	error
	StatusCode() int   // actual HTTP status code
	ErrorCode() string // error code returned in response body from CC or UAA
}

type baseHttpError struct {
	statusCode   int
	apiErrorCode string
	description  string
}

type HttpNotFoundError struct {
	baseHttpError
}

func NewHttpError(statusCode int, code string, description string) error {
	err := baseHttpError{
		statusCode:   statusCode,
		apiErrorCode: code,
		description:  description,
	}
	switch statusCode {
	case 404:
		return &HttpNotFoundError{err}
	default:
		return &err
	}
}

func (err *baseHttpError) StatusCode() int {
	return err.statusCode
}

func (err *baseHttpError) Error() string {
	return fmt.Sprintf(
		"Server error, status code: %d, error code: %s, message: %s",
		err.statusCode,
		err.apiErrorCode,
		err.description,
	)
}

func (err *baseHttpError) ErrorCode() string {
	return err.apiErrorCode
}

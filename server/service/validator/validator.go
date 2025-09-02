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

package validator

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/apache/servicecomb-service-center/server/config"
)

type CustomValidator interface {
	Validate(v interface{}) (bool, error)
}

var customValidators = map[string]CustomValidator{}

func registerCustomValidator(name string, validator CustomValidator) {
	customValidators[name] = validator
}

func baseCheck(v interface{}) error {
	if v == nil {
		return errors.New("data is nil")
	}
	sv := reflect.ValueOf(v)
	if sv.Kind() == reflect.Ptr && sv.IsNil() {
		return errors.New("pointer is nil")
	}
	return nil
}

// customValidate 自定义校验，校验器不存在或校验失败返回err
// v: 待校验数据
// targetValidators: 待使用的自定义校验器
func customValidate(v interface{}, targetValidators ...string) error {
	if !config.GetServer().EnableCustomValidate {
		return nil
	}
	for _, validatorName := range targetValidators {
		validator, exists := customValidators[validatorName]
		if !exists {
			return errors.New(fmt.Sprintf("validator:%s is not registered", validatorName))
		}
		validate, err := validator.Validate(v)
		if err != nil {
			return err
		}
		if !validate {
			return errors.New(fmt.Sprintf("Validate failed,validator:%s", validatorName))
		}
	}
	return nil
}

func initCustomValidator() {
	if !config.GetServer().EnableCustomValidate {
		return
	}
	initPasswordCustomValidator()
}

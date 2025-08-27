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

	"github.com/apache/servicecomb-service-center/pkg/log"
)

type CustomValidator interface {
	validate(v interface{}) (bool, error)
}

var customValidators = map[string]CustomValidator{}

func RegisterCustomValidator(name string, validator CustomValidator) {
	customValidators[name] = validator
}

// 定义一个具体的类型
type StringValidator struct{}

// 实现 CustomValidator[string] 接口
func (sv *StringValidator) validate(v interface{}) (bool, error) {
	// 简单的验证逻辑，例如检查字符串长度
	return len(v.(string)) > 0, nil
}

func baseCheck(v interface{}) error {
	RegisterCustomValidator("StringValidator", &StringValidator{})
	if v == nil {
		return errors.New("data is nil")
	}
	sv := reflect.ValueOf(v)
	if sv.Kind() == reflect.Ptr && sv.IsNil() {
		return errors.New("pointer is nil")
	}
	return nil
}

func customValidate(v interface{}, targetValidators ...string) error {
	for _, validatorName := range targetValidators {
		validator, exists := customValidators[validatorName]
		if !exists {
			log.Info(fmt.Sprintf("validator:%s is not registered,skip", validatorName))
			continue
		}
		validate, err := validator.validate(v)
		if err != nil {
			return err
		}
		if !validate {
			return errors.New(fmt.Sprintf("validate failed,validator:%s", validatorName))
		}
		return nil
	}

	return nil
}

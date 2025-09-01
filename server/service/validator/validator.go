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

func customValidate(v interface{}, targetValidators ...string) error {
	if len(customValidators) == 0 {
		return nil
	}
	for _, validatorName := range targetValidators {
		validator, exists := customValidators[validatorName]
		if !exists {
			log.Info(fmt.Sprintf("validator:%s is not registered,skip", validatorName))
			continue
		}
		validate, err := validator.Validate(v)
		if err != nil {
			return err
		}
		if !validate {
			return errors.New(fmt.Sprintf("Validate failed,validator:%s", validatorName))
		}
		return nil
	}
	return nil
}

func initCustomValidator() {
	if !config.GetServer().EnableCustomValidate {
		return
	}
	initPasswordCustomValidator()
}

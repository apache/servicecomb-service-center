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
	"bufio"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"

	"github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/pkg/validate"
)

var createAccountValidator = &validate.Validator{}
var updateAccountValidator = &validate.Validator{}
var createRoleValidator = &validate.Validator{}
var batchCreateAccountsRequestValidator = &validate.Validator{}
var changePWDValidator = &validate.Validator{}
var accountLoginValidator = &validate.Validator{}

var PasswordCustomValidator = "passwordCustomValidator"

func init() {
	createAccountValidator.AddRule("Name", &validate.Rule{Min: 1, Max: 64, Regexp: accountNameRegex})
	createAccountValidator.AddRule("Roles", &validate.Rule{Min: 1, Max: 5, Regexp: nameRegex})
	createAccountValidator.AddRule("Password", &validate.Rule{Regexp: &validate.PasswordChecker{}})
	createAccountValidator.AddRule("Status", &validate.Rule{Regexp: accountStatusRegex})

	batchCreateAccountsRequestValidator.AddRule("Accounts", &validate.Rule{Min: 1, Max: 20})

	updateAccountValidator.AddRule("Roles", createAccountValidator.GetRule("Roles"))
	updateAccountValidator.AddRule("Status", createAccountValidator.GetRule("Status"))

	createRoleValidator.AddRule("Name", &validate.Rule{Min: 1, Max: 64, Regexp: nameRegex})

	changePWDValidator.AddRule("Password", &validate.Rule{Regexp: &validate.PasswordChecker{}})
	changePWDValidator.AddRule("Name", &validate.Rule{Regexp: accountNameRegex})

	accountLoginValidator.AddRule("TokenExpirationTime", &validate.Rule{Regexp: &validate.TokenExpirationTimeChecker{}})

	initCustomValidator()
}

func initPasswordCustomValidator() {
	weakPasswordPath := filepath.Join(util.GetAppRoot(), "conf", "weakpassord.txt")
	weakPasswords, err := loadWeakPasswords(weakPasswordPath)
	if err != nil {
		log.Error("failed to load weak password", err)
		panic(err)
	}
	registerCustomValidator(PasswordCustomValidator, &passwordValidator{weakPasswords: weakPasswords})
}

type passwordValidator struct {
	weakPasswords map[string]struct{}
}

func (pv *passwordValidator) Validate(v interface{}) (bool, error) {
	account := v.(*rbac.Account)
	_, exist := pv.weakPasswords[account.Password]
	if exist {
		return false, errors.New("the password is a weak password")
	}
	return true, nil
}

// loadWeakPasswords loads the weak passwords from the file, decodes them from Base64, and stores them in a map.
func loadWeakPasswords(filePath string) (map[string]struct{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	weakPasswords := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		encodedPassword := scanner.Text()
		decodedPassword, err := base64.StdEncoding.DecodeString(encodedPassword)
		if err != nil {
			return nil, err
		}
		weakPasswords[string(decodedPassword)] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return weakPasswords, nil
}

func ValidateCreateAccount(a *rbac.Account) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	err = customValidate(a, PasswordCustomValidator)
	if err != nil {
		return err
	}
	return createAccountValidator.Validate(a)
}
func ValidateBatchCreateAccountsRequest(a *rbac.BatchCreateAccountsRequest) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return batchCreateAccountsRequestValidator.Validate(a)
}
func ValidateUpdateAccount(a *rbac.Account) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return updateAccountValidator.Validate(a)
}
func ValidateCreateRole(a *rbac.Role) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return createRoleValidator.Validate(a)
}
func ValidateAccountLogin(a *rbac.Account) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return accountLoginValidator.Validate(a)
}

func ValidateChangePWD(a *rbac.Account) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	err = customValidate(a, PasswordCustomValidator)
	if err != nil {
		return err
	}
	return changePWDValidator.Validate(a)
}

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

package privacy

import (
	"strings"

	scrypt "github.com/elithrar/simple-scrypt"
	"golang.org/x/crypto/bcrypt"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

const (
	algBcrypt = "$2a$"
)

var ScryptParams = scrypt.Params{N: 1024, R: 8, P: 1, SaltLen: 8, DKLen: 32}

// DefaultManager default manager
var DefaultManager PasswordManager = &passwordManager{}

type PasswordManager interface {
	EncryptPassword(pwd string) (string, error)
	CheckPassword(hashedPwd, pwd string) bool
}

type passwordManager struct {
}

func (p *passwordManager) EncryptPassword(pwd string) (string, error) {
	hash, err := scrypt.GenerateFromPassword([]byte(pwd), ScryptParams)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (p *passwordManager) CheckPassword(hashedPwd, pwd string) bool {
	if strings.HasPrefix(hashedPwd, algBcrypt) {
		err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(pwd))
		if err == bcrypt.ErrMismatchedHashAndPassword {
			log.Warn("incorrect password attempts")
		}
		return err == nil
	}
	err := scrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(pwd))
	if err == scrypt.ErrMismatchedHashAndPassword {
		log.Warn("incorrect password attempts")
	}
	return err == nil
}

func ScryptPassword(pwd string) (string, error) {
	return DefaultManager.EncryptPassword(pwd)
}
func SamePassword(hashedPwd, pwd string) bool {
	return DefaultManager.CheckPassword(hashedPwd, pwd)
}

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
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/elithrar/simple-scrypt"
	"github.com/go-chassis/foundation/stringutil"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

const (
	algBcrypt = "$2a$"
)

//HashPassword
//Deprecated: use ScryptPassword, this is only for unit test to test compatible with old version
func HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 14)
	if err != nil {
		return "", err
	}
	return stringutil.Bytes2str(hash), nil
}
func ScryptPassword(pwd string) (string, error) {
	hash, err := scrypt.GenerateFromPassword([]byte(pwd), scrypt.DefaultParams)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
func SamePassword(hashedPwd, pwd string) bool {
	if strings.HasPrefix(hashedPwd, algBcrypt) {
		err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(pwd))
		if err == bcrypt.ErrMismatchedHashAndPassword {
			log.Warn("incorrect password attempts")
		}
		return err == nil
	}
	err := scrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(pwd))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		log.Warn("incorrect password attempts")
	}
	return err == nil

}

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

package rbac

import (
	"context"
	"crypto/rsa"
	"errors"
	"io/ioutil"

	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/security/authr"
	"github.com/go-chassis/go-chassis/v2/security/secret"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/plugin/security/cipher"
)

const (
	RootName     = "root"
	InitPassword = "SC_INIT_ROOT_PASSWORD"
)

var (
	ErrEmptyCurrentPassword = errors.New("current password should not be empty")
	ErrNoPermChangeAccount  = errors.New("can not change other account password")
	ErrWrongPassword        = errors.New("current pwd is wrong")
	ErrSamePassword         = errors.New("the password can not be same as old one")
)

//Init decide whether enable rbac function and save the build-in roles to db
func Init() {
	if !Enabled() {
		log.Info("rbac is disabled")
		return
	}
	InitResourceMap()
	err := authr.Init()
	if err != nil {
		log.Fatal("can not enable auth module", err)
	}
	// build-in role init
	initBuildInRole()
	initBuildInAccount()
	add2WhiteAPIList()
	readPrivateKey()
	readPublicKey()
	log.Info("rbac is enabled")
}

func add2WhiteAPIList() {
	rbac.Add2WhiteAPIList(APITokenGranter)
	rbac.Add2WhiteAPIList("/v4/:project/registry/version", "/version")
	rbac.Add2WhiteAPIList("/v4/:project/registry/health", "/health")

	// user can list self permission without account get permission
	Add2CheckPermWhiteAPIList(APISelfPerms)
	// user can change self password without account modify permission
	Add2CheckPermWhiteAPIList(APIAccountPassword)
}

func initBuildInAccount() {
	accountExist, err := AccountExist(context.Background(), RootName)
	if err != nil {
		log.Fatal("can not enable auth module", err)
	}
	if !accountExist {
		initFirstTime()
	}
}

//read key to memory
func readPrivateKey() {
	pf := config.GetString("rbac.privateKeyFile", "", config.WithStandby("rbac_rsa_private_key_file"))
	// 打开文件
	data, err := ioutil.ReadFile(pf)
	if err != nil {
		log.Fatal("can not read private key", err)
		return
	}
	err = archaius.Set("rbac_private_key", string(data))
	if err != nil {
		log.Fatal("can not init rbac", err)
		return
	}
	log.Info("read private key success")
}

//read key to memory
func readPublicKey() {
	pf := config.GetString("rbac.publicKeyFile", "", config.WithStandby("rbac_rsa_public_key_file"))
	// 打开文件
	content, err := ioutil.ReadFile(pf)
	if err != nil {
		log.Fatal("can not find public key", err)
		return
	}
	err = archaius.Set("rbac_public_key", string(content))
	if err != nil {
		log.Fatal("", err)
	}
	log.Info("read public key success")
}
func initFirstTime() {
	//handle root account
	pwd := getPassword()
	if len(pwd) == 0 {
		log.Warn("skip init root account! Cause by " + InitPassword + " is empty. " +
			"Please use the private key to generate a ROOT token and call " + APIAccountList + " create ROOT!")
		return
	}
	a := &rbac.Account{
		Name:     RootName,
		Roles:    []string{rbac.RoleAdmin},
		Password: pwd,
	}
	err := CreateAccount(context.Background(), a)
	if err == nil {
		log.Info("root account init success")
		return
	}
	if errsvc.IsErrEqualCode(err, rbac.ErrAccountConflict) {
		log.Info("root account already exist")
		return
	}
	log.Fatal("can not enable rbac, init root account failed", err)
}

func getPassword() string {
	p := archaius.GetString(InitPassword, "")
	if p == "" {
		return ""
	}
	d, err := cipher.Decrypt(p)
	if err != nil {
		log.Warn("cipher fallback: " + err.Error())
		return p
	}
	return d
}

func Enabled() bool {
	return config.GetRBAC().EnableRBAC
}

//PublicKey get public key to verify a token
func PublicKey() string {
	return archaius.GetString("rbac_public_key", "")
}

//privateKey get decrypted private key to verify a token
func privateKey() string {
	ep := archaius.GetString("rbac_private_key", "")
	p, err := cipher.Decrypt(ep)
	if err != nil {
		log.Warn("cipher fallback: " + err.Error())
		return ep
	}
	return p
}

//GetPrivateKey return rsa key instance
func GetPrivateKey() (*rsa.PrivateKey, error) {
	sk := privateKey()
	p, err := secret.ParseRSAPrivateKey(sk)
	if err != nil {
		log.Error("can not get key:", err)
		return nil, err
	}
	return p, nil
}

//MakeBanKey return ban key
func MakeBanKey(name, ip string) string {
	return name + "::" + ip
}

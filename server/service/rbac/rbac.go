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

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/service"
	"github.com/apache/servicecomb-service-center/server/service/cipher"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	"github.com/astaxie/beego"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/security/authr"
	"github.com/go-chassis/go-chassis/security/secret"
)

const (
	RootName     = "root"
	InitPassword = "SC_INIT_ROOT_PASSWORD"
	PubFilePath  = "rbac_rsa_public_key_file"
)
const (
	ResourceAccount = "account"
)

var (
	ErrEmptyCurrentPassword = errors.New("current password should not be empty")
	ErrNoPermChangeAccount  = errors.New("can not change other account password")
	ErrWrongPassword        = errors.New("current pwd is wrong")
	ErrSamePassword         = errors.New("the password can not be same as old one")
	ErrEmptyPassword        = errors.New("empty password")
)

//Init decide whether enable rbac function and save root account to db
// if db has root account, abort creating.
func Init() {
	if !Enabled() {
		log.Info("rbac is disabled")
		return
	}
	initResourceMap()
	err := authr.Init()
	if err != nil {
		log.Fatal("can not enable auth module", err)
	}
	accountExist, err := dao.AccountExist(context.Background(), RootName)
	if err != nil {
		log.Fatal("can not enable auth module", err)
	}
	if !accountExist {
		initFirstTime(RootName)
	}
	readPrivateKey()
	readPublicKey()
	rbacframe.Add2WhiteAPIList("/v4/token")
	core.ServerInfo.Config.EnableRBAC = true
	log.Info("rbac is enabled")
}
func initResourceMap() {
	rbacframe.MapResource("/v4/account", ResourceAccount)
	rbacframe.MapResource("/v4/account/:name", ResourceAccount)
	rbacframe.MapResource("/v4/role", "role")
	//TODO now simply write dead code "*" to map all other API except account and role to service, should define resource for every API in future
	rbacframe.MapResource("*", "service")
}

//readPublicKey read key to memory
func readPrivateKey() {
	pf := beego.AppConfig.String("rbac_rsa_private_key_file")
	// 打开文件
	data, err := ioutil.ReadFile(pf)
	if err != nil {
		log.Fatal("can not read private key", err)
		return
	}
	err = archaius.Set("rbac_private_key", string(data))
	if err != nil {
		log.Fatal("can not init rbac", err)
	}
	log.Info("read private key success")
}

//readPublicKey read key to memory
func readPublicKey() {
	pf := beego.AppConfig.String(PubFilePath)
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
func initFirstTime(admin string) {
	//handle root account
	pwd, err := getPassword()
	if err != nil {
		log.Fatal("can not enable rbac, password is empty", nil)
	}
	a := &rbacframe.Account{
		Name:     admin,
		Password: pwd,
		Role:     rbacframe.RoleAdmin,
	}
	err = service.ValidateCreateAccount(a)
	if err != nil {
		log.Fatal("invalid pwd", err)
		return
	}
	if err := dao.CreateAccount(context.Background(), a); err != nil {
		if err == dao.ErrDuplicated {
			log.Info("rbac is enabled")
			return
		}
		log.Fatal("can not enable rbac, init root account failed", err)
	}
	log.Info("root account init success")
}

func getPassword() (string, error) {
	p := archaius.GetString(InitPassword, "")
	if p == "" {
		log.Fatal("can not enable rbac, password is empty", nil)
		return "", ErrEmptyPassword
	}
	d, err := cipher.Decrypt(p)
	if err != nil {
		log.Warn("cipher fallback: " + err.Error())
		return p, nil
	}
	return d, nil
}

func Enabled() bool {
	return beego.AppConfig.DefaultBool("rbac_enabled", false)
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

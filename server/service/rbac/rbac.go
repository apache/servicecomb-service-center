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
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/model"
	"github.com/apache/servicecomb-service-center/server/service/cipher"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	"github.com/astaxie/beego"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/security/authr"
	"github.com/go-chassis/go-chassis/security/secret"
	"io"
	"os"
)

const (
	InitRoot     = "SC_INIT_ROOT_USERNAME"
	InitPassword = "SC_INIT_ROOT_PASSWORD"
	InitPrivate  = "SC_INIT_PRIVATE_KEY"
)

//Init decide whether enable rbac function and save root account to db
// if db has root account, abort creating.
func Init() {
	if !Enabled() {
		log.Info("rbac is disabled")
		return
	}
	err := authr.Init()
	if err != nil {
		log.Fatal("can not enable auth module", err)
	}
	admin := archaius.GetString(InitRoot, "")
	if admin == "" {
		log.Fatal("can not enable rbac, root is empty", nil)
		return
	}
	accountExist, err := dao.AccountExist(context.Background(), admin)
	if err != nil {
		log.Fatal("can not enable auth module", err)
	}
	if !accountExist {
		initFirstTime(admin)
	}
	overrideSecretKey()
	readPublicKey()
	log.Info("rbac is enabled")
}

//readPublicKey read key to memory
func readPublicKey() {
	pf := beego.AppConfig.String("rbac_rsa_pub_key_file")
	// 打开文件
	fp, err := os.Open(pf)
	if err != nil {
		log.Fatal("can not find public key", err)
		return
	}
	defer fp.Close()
	buf := make([]byte, 1024)
	for {
		// 循环读取文件
		_, err := fp.Read(buf)
		if err == io.EOF { // io.EOF表示文件末尾
			break
		}

	}
	archaius.Set("rbac_public_key", string(buf))
}
func initFirstTime(admin string) {
	//handle root account
	pwd := archaius.GetString(InitPassword, "")
	if pwd == "" {
		log.Fatal("can not enable rbac, password is empty", nil)
	}
	pwd, err := cipher.Encrypt(pwd)
	if err != nil {
		log.Fatal("can not enable rbac, encryption failed", err)
	}
	if err := dao.CreateAccount(context.Background(), &model.Account{
		Name:     admin,
		Password: pwd,
	}); err != nil {
		if err == dao.ErrDuplicated {
			log.Info("rbac is enabled")
			return
		}
		log.Fatal("can not enable rbac, init root account failed", err)
	}
	log.Info("root account init success")
}

//should override key on each start procedure,
//so that a system such as kubernetes can use secret to distribute a new secret to revoke the old one
func overrideSecretKey() {
	secret := archaius.GetString(InitPrivate, "")
	if secret == "" {
		log.Fatal("can not enable rbac, secret is empty", nil)
	}
	if err := dao.OverrideSecret(context.Background(), secret); err != nil {
		log.Fatal("can not save secret", err)
	}

}
func Enabled() bool {
	return beego.AppConfig.DefaultBool("rbac_enabled", false)
}

//PublicKey get public key to verify a token
func PublicKey() string {
	return archaius.GetString("rbac_public_key", "")
}

//GetSecretStr return decrypted secret
func GetSecretStr(ctx context.Context) (string, error) {
	sk, err := dao.GetSecret(ctx)
	if err != nil {
		return "", err
	}
	skStr, err := cipher.Decrypt(string(sk))
	if err != nil {
		log.Error("can not decrypt:", err)
		return "", err
	}
	return skStr, nil
}

//GetPrivateKey return rsa key instance
func GetPrivateKey(ctx context.Context) (*rsa.PrivateKey, error) {
	sk, err := GetSecretStr(ctx)
	if err != nil {
		return nil, err
	}
	p, err := secret.ParseRSAPrivateKey(sk)
	if err != nil {
		log.Error("can not get key:", err)
		return nil, err
	}
	return p, nil
}

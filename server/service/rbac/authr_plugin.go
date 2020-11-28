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
	"errors"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/apache/servicecomb-service-center/server/service/rbac/dao"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chassis/go-chassis/v2/security/authr"
	"github.com/go-chassis/go-chassis/v2/security/token"
)

var ErrUnauthorized = errors.New("wrong user name or password")

//EmbeddedAuthenticator is sc default auth plugin, RBAC data is persisted in etcd
type EmbeddedAuthenticator struct {
}

func newEmbeddedAuthenticator(opts *authr.Options) (authr.Authenticator, error) {
	return &EmbeddedAuthenticator{}, nil
}

//Login check db user and password,will verify and return token for valid account
func (a *EmbeddedAuthenticator) Login(ctx context.Context, user string, password string, opts ...authr.LoginOption) (string, error) {
	opt := &authr.LoginOptions{}
	for _, o := range opts {
		o(opt)
	}
	exist, err := dao.AccountExist(ctx, user)
	if err != nil {
		log.Error("check account err", err)
		return "", err
	}
	if !exist {
		return "", ErrUnauthorized
	}
	account, err := dao.GetAccount(ctx, user)
	if err != nil {
		log.Error("get account err", err)
		return "", err
	}
	same := SamePassword(account.Password, password)
	if user == account.Name && same {
		secret, err := GetPrivateKey()
		if err != nil {
			return "", err
		}
		tokenStr, err := token.Sign(map[string]interface{}{
			rbacframe.ClaimsUser:  user,
			rbacframe.ClaimsRoles: account.Roles,
		},
			secret,
			token.WithExpTime(opt.ExpireAfter),
			token.WithSigningMethod(token.RS512)) //TODO config for each user
		if err != nil {
			log.Errorf(err, "can not sign a token")
			return "", err
		}
		return tokenStr, nil
	}
	return "", ErrUnauthorized
}
func (a *EmbeddedAuthenticator) Authenticate(ctx context.Context, tokenStr string) (interface{}, error) {
	p, err := jwt.ParseRSAPublicKeyFromPEM([]byte(PublicKey()))
	if err != nil {
		log.Error("can not parse public key", err)
		return nil, err
	}
	return rbacframe.Authenticate(tokenStr, p)
}

func init() {
	authr.Install("default", newEmbeddedAuthenticator)
}

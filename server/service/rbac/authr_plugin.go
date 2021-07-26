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
	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-chassis/v2/security/authr"
	"github.com/go-chassis/go-chassis/v2/security/token"
)

//EmbeddedAuthenticator is sc default auth plugin, RBAC data is persisted in etcd
type EmbeddedAuthenticator struct {
}

func newEmbeddedAuthenticator(opts *authr.Options) (authr.Authenticator, error) {
	return &EmbeddedAuthenticator{}, nil
}

//Login check db user and password,will verify and return token for valid account
func (a *EmbeddedAuthenticator) Login(ctx context.Context, user string, password string, opts ...authr.LoginOption) (string, error) {
	ip := util.GetIPFromContext(ctx)
	if IsBanned(MakeBanKey(user, ip)) {
		log.Warnf("ip [%s] is banned, account: %s", ip, user)
		return "", ErrAccountBlocked
	}
	opt := &authr.LoginOptions{}
	for _, o := range opts {
		o(opt)
	}
	account, err := GetAccount(ctx, user)
	if err != nil {
		if errsvc.IsErrEqualCode(err, rbac.ErrAccountNotExist) {
			TryLockAccount(MakeBanKey(user, ip))
			return "", UserOrPwdWrongError()
		}
		return "", err
	}
	same := privacy.SamePassword(account.Password, password)
	if !same {
		TryLockAccount(MakeBanKey(user, ip))
		return "", UserOrPwdWrongError()
	}

	secret, err := GetPrivateKey()
	if err != nil {
		return "", err
	}
	tokenStr, err := token.Sign(map[string]interface{}{
		rbac.ClaimsUser:  user,
		rbac.ClaimsRoles: account.Roles,
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

//Authenticate parse a token to claims
func (a *EmbeddedAuthenticator) Authenticate(ctx context.Context, tokenStr string) (interface{}, error) {
	p, err := jwt.ParseRSAPublicKeyFromPEM([]byte(PublicKey()))
	if err != nil {
		log.Error("can not parse public key", err)
		return nil, err
	}
	claims, err := a.authToken(tokenStr, p)
	if err != nil {
		if a.isTokenExpiredError(err) {
			return nil, ErrTokenExpired
		}
		return nil, err
	}
	return claims, nil
}

func (a *EmbeddedAuthenticator) isTokenExpiredError(err error) bool {
	if err == nil {
		return false
	}
	vErr, ok := err.(*jwt.ValidationError)
	if !ok {
		return false
	}
	if vErr.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
		return true
	}
	return false
}

func (a *EmbeddedAuthenticator) authToken(tokenStr string, pub *rsa.PublicKey) (map[string]interface{}, error) {
	return token.Verify(tokenStr, func(claims interface{}, method token.SigningMethod) (interface{}, error) {
		return pub, nil
	})
}

func UserOrPwdWrongError() error {
	if AuthResource(ResourceService) {
		return ErrUserOrPwdWrong
	}
	return ErrUserOrPwdWrongEx
}

func init() {
	authr.Install("default", newEmbeddedAuthenticator)
}

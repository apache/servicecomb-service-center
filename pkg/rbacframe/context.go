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

package rbacframe

import (
	"context"
	"k8s.io/apimachinery/pkg/util/sets"
)

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key string

// accountKey is the key for user.User values in Contexts. It is
// unexported; clients use user.NewContext and user.FromContext
// instead of using this key directly.
var accountKey key

// NewContext returns a new Context that carries value claims.
func NewContext(ctx context.Context, claims interface{}) context.Context {
	return context.WithValue(ctx, accountKey, claims)
}

// FromContext returns the account value stored in ctx, if any.
func FromContext(ctx context.Context) interface{} {
	a := ctx.Value(accountKey)
	return a
}

var whiteAPIList = sets.NewString()

func Add2WhiteAPIList(path ...string) {
	whiteAPIList.Insert(path...)
}
func MustAuth(url string) bool {
	return !whiteAPIList.Has(url)
}

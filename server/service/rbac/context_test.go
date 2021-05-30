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

package rbac_test

import (
	"context"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
	"testing"
)

func TestUserFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"have a claims with user A, should return A",
			args{
				ctx: context.WithValue(context.Background(), rbacsvc.CtxRequestClaims, map[string]interface{}{"account": "A"}),
			},
			"A",
		},
		{
			"no claims, should return empty",
			args{
				ctx: context.Background(),
			},
			"",
		},
		{
			"have invalid claims, should return empty",
			args{
				ctx: context.WithValue(context.Background(), rbacsvc.CtxRequestClaims, nil),
			},
			"",
		},
		{
			"have a claims with no user, should return empty",
			args{
				ctx: context.WithValue(context.Background(), rbacsvc.CtxRequestClaims, map[string]interface{}{}),
			},
			"",
		},
		{
			"have a claims with invalid user, should return empty",
			args{
				ctx: context.WithValue(context.Background(), rbacsvc.CtxRequestClaims, map[string]interface{}{"account": nil}),
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rbacsvc.UserFromContext(tt.args.ctx); got != tt.want {
				t.Errorf("UserFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

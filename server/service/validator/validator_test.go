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

package validator_test

import (
	"github.com/apache/servicecomb-service-center/server/service/validator"
	"testing"

	"github.com/go-chassis/cari/rbac"
)

func TestValidateCreateAccount(t *testing.T) {
	type args struct {
		a *rbac.Account
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "given valid field",
			args: args{a: &rbac.Account{
				Name:     "tester-_1",
				Password: "Pwd0000_1",
				Roles:    []string{"admin", "developer"},
			}},
			wantErr: false,
		},
		{name: "given invalid account name",
			args: args{a: &rbac.Account{
				Name:     "tester*",
				Password: "Pwd0000_1",
				Roles:    []string{"admin-2", "developer", "admin-a", "admin.a"},
			}},
			wantErr: true,
		},
		{name: "given invalid account name",
			args: args{a: &rbac.Account{
				Name:     "tester*a",
				Password: "Pwd0000_1",
				Roles:    []string{"admin", "developer"},
			}},
			wantErr: true,
		},
		{name: "given invalid role name",
			args: args{a: &rbac.Account{
				Name:     "tester",
				Password: "Pwd0000_1",
				Roles:    []string{"adm*in", "developer"},
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if err := validator.ValidateCreateAccount(tt.args.a); (err != nil) != tt.wantErr {
			t.Errorf("%q. ValidateCreateAccount() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

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

package datasource

import (
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
)

const (
	SPLIT                 = "/"
	ServiceKeyPrefix      = "/cse-sr/ms/files"
	InstanceKeyPrefix     = "/cse-sr/inst/files"
	RegistryDomain        = "default"
	RegistryProject       = "default"
	RegistryDomainProject = "default/default"
	RegistryAppID         = "default"
	Provider              = "p"

	ResourceAccount = "account"
	ResourceRole    = "role"
)

// WrapErrResponse is temp func here to wait finish to refact the discosvc pkg
func WrapErrResponse(respErr error) (*pb.Response, error) {
	err, ok := respErr.(*errsvc.Error)
	if !ok {
		return pb.CreateResponse(pb.ErrInternal, err.Error()), err
	}
	resp := pb.CreateResponseWithSCErr(err)
	if err.InternalError() {
		return resp, err
	}
	return resp, nil
}

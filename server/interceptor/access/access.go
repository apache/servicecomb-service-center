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

package access

import (
	"fmt"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/pkg/validate"
	"github.com/apache/servicecomb-service-center/server/core"
)

var (
	serverName string
)

func init() {
	serverName = core.Service.ServiceName + "/" + core.Service.Version
}

func Intercept(w http.ResponseWriter, r *http.Request) error {
	w.Header().Add(rest.HeaderServer, serverName)

	if !validate.IsRequestURI(r.RequestURI) {
		err := fmt.Errorf("Invalid Request URI %s", r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(util.StringToBytesWithNoCopy(err.Error()))
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

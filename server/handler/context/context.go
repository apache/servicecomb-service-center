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

package context

import (
	"github.com/apache/servicecomb-service-center/pkg/chain"
	roa "github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"net/http"
)

type Handler struct {
}

func (c *Handler) Handle(i *chain.Invocation) {
	var (
		err     error
		v3      v3Context
		v4      v4Context
		r       = i.Context().Value(roa.CtxRequest).(*http.Request)
		pattern = i.Context().Value(roa.CtxMatchPattern).(string)
	)

	switch {
	case util.IsVersionOrHealthPattern(pattern):
	case v3.IsMatch(r):
		err = v3.Do(r)
	case v4.IsMatch(r):
		err = v4.Do(r)
	}

	if err != nil {
		i.Fail(err)
		return
	}

	i.WithContext("x-remote-ip", util.GetRealIP(r))

	i.Next()
}

func RegisterHandlers() {
	chain.RegisterHandler(roa.ServerChainName, &Handler{})
}

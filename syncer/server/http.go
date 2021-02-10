/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package server

import (
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/go-chassis/v2"
	rf "github.com/go-chassis/go-chassis/v2/server/restful"
)

const (
	Message = "Deliver full synchronization task success!"
)

func (s *Server) FullSync(b *rf.Context) {
	s.mux.Lock()
	s.triggered = true
	s.mux.Unlock()
	err := b.Write([]byte(Message))
	if err != nil {
		log.Error("", err)
	}
}

func (s *Server) URLPatterns() []rf.Route {
	return []rf.Route{
		{Method: http.MethodGet, Path: "/v1/syncer/full-synchronization", ResourceFunc: s.FullSync},
	}
}

//if you use go run main.go instead of binary run, plz export CHASSIS_HOME=/{path}/{to}/server/

func (s *Server) NewHTTPServer() {
	chassis.RegisterSchema("rest", s)
	if err := chassis.Init(); err != nil {
		log.Error("Init failed.", err)
		return
	}
	err := chassis.Run()
	if err != nil {
		log.Error("fail to run http server", err)
	}
}

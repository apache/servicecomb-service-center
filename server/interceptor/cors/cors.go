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

package cors

import (
	"errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/rs/cors"
	"net/http"
)

var CORS *cors.Cors

func init() {
	CORS = cors.New(cors.Options{
		AllowedHeaders: []string{"Origin", "Accept", "Content-Type", "X-Domain-Name", "X-ConsumerId"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "UPDATE"},
	})
}

func Intercept(w http.ResponseWriter, r *http.Request) (err error) {
	CORS.HandlerFunc(w, r)
	if r.Method == "OPTIONS" {
		log.Debugf("identify the current request is a CORS, url: %s", r.RequestURI)
		err = errors.New("Handle the preflight request")
	}
	return
}

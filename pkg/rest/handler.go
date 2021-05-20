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

package rest

import (
	"net/http"
	"net/url"
)

type urlPatternHandler struct {
	Name string
	Path string
	http.Handler
}

func (roa *urlPatternHandler) try(path string) (p string, _ bool) {
	var i, j int
	l, sl := len(roa.Path), len(path)
	for i < sl {
		switch {
		case j >= l:
			if roa.Path != "/" && l > 0 && roa.Path[l-1] == '/' {
				return p, true
			}
			return "", false
		case roa.Path[j] == ':':
			var val string
			var nextc byte
			o := j
			_, nextc, j = match(roa.Path, isAlnum, 0, j+1)
			val, _, i = match(path, matchParticipial, nextc, i)

			p += url.QueryEscape(roa.Path[o:j]) + "=" + url.QueryEscape(val) + "&"
		case path[i] == roa.Path[j]:
			i++
			j++
		default:
			return "", false
		}
	}
	if j != l {
		return "", false
	}
	return p, true
}

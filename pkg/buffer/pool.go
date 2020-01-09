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
package buffer

import (
	"bytes"
	"sync"
)

type Pool struct {
	pool sync.Pool
}

func (p *Pool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

func (p *Pool) Put(buf *bytes.Buffer) {
	buf.Reset()
	p.pool.Put(buf)
}

func NewPool(s int) *Pool {
	return &Pool{
		pool: sync.Pool{
			New: func() interface{} {
				b := bytes.NewBuffer(make([]byte, s))
				b.Reset()
				return b
			},
		},
	}
}

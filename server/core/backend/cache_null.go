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
package backend

var (
	NullCache  = &nullCache{}
	NullCacher = &nullCacher{}
)

type nullCache struct {
}

func (n *nullCache) Revision() int64 {
	return 0
}

func (n *nullCache) Data(interface{}) interface{} {
	return nil
}

func (n *nullCache) Have(interface{}) bool {
	return false
}

func (n *nullCache) Size() int {
	return 0
}

type nullCacher struct {
}

func (n *nullCacher) Name() string {
	return ""
}

func (n *nullCacher) Cache() Cache {
	return NullCache
}

func (n *nullCacher) Run() {}

func (n *nullCacher) Stop() {}

func (n *nullCacher) Ready() <-chan struct{} {
	return closedCh
}

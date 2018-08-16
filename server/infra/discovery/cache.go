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
package discovery

type Cache interface {
	Name() string
	Size() int // the bytes size of the cache

	Get(k string) *KeyValue
	GetAll(arr *[]*KeyValue) int
	GetPrefix(prefix string, arr *[]*KeyValue) int
	ForEach(iter func(k string, v *KeyValue) (next bool))

	Put(k string, v *KeyValue)
	Remove(k string)
}

type Cacher interface {
	Cache() Cache
}

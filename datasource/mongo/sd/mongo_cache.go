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

package sd

type MongoCache interface {
	MongoCacheReader
	Dirty() bool
	Clear()
	MarkDirty()
	ProcessUpdate(event MongoEvent)
	ProcessDelete(event MongoEvent)
	isValueNotUpdated(value interface{}, newValue interface{}) bool
}

type MongoCacheReader interface {
	Name() string // The name of implementation

	Size() int // the bytes size of the cache

	// Get gets a value by docid
	Get(id string) interface{}

	// GetValue gets a value by index
	GetValue(index string) []interface{}

	// ForEach executes the given function for each of the k-v
	ForEach(iter func(k string, v interface{}) (next bool))
}

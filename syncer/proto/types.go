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

package proto

type SyncMapping []*MappingEntry

func (s SyncMapping) OriginIndex(instanceID string) int {
	for index, val := range s {
		if val.OrgInstanceID == instanceID {
			return index
		}
	}
	return -1
}

func (s SyncMapping) CurrentIndex(instanceID string) int {
	for index, val := range s {
		if val.CurInstanceID == instanceID {
			return index
		}
	}
	return -1
}

type Expansions []*Expansion

func (e Expansions) Find(kind string, labels map[string]string) Expansions {
	matches := make(Expansions, 0, 10)
	for _, expansion := range e {
		if expansion.Kind != kind || !matchLabels(expansion.Labels, labels) {
			continue
		}
		matches = append(matches, expansion)
	}
	return matches
}

func matchLabels(src, dst map[string]string) bool {
	for key, val := range dst {
		if src[key] != val {
			return false
		}
	}
	return true
}

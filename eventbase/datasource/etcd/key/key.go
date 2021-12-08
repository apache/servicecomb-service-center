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

package key

import (
	"strconv"
	"strings"
)

const (
	split     = "/"
	syncer    = "syncer"
	task      = "task"
	tombstone = "tombstone"
)

func getSyncRootKey() string {
	return split + syncer + split + task
}

func getTombstoneRootKey() string {
	return split + tombstone
}

func TaskKey(domain, project, taskID string, timestamp int64) string {
	strTimestamp := strconv.FormatInt(timestamp, 10)
	return strings.Join([]string{getSyncRootKey(), domain, project, strTimestamp, taskID}, split)
}

func TaskList(domain, project string) string {
	if len(project) == 0 {
		return strings.Join([]string{getSyncRootKey(), domain, ""}, split)
	}
	return strings.Join([]string{getSyncRootKey(), domain, project, ""}, split)
}

func TombstoneList(domain, project string) string {
	if len(project) == 0 {
		return strings.Join([]string{getTombstoneRootKey(), domain, ""}, split)
	}
	return strings.Join([]string{getTombstoneRootKey(), domain, project, ""}, split)
}

func TombstoneKey(domain, project, resourceType, resourceID string) string {
	return strings.Join([]string{getTombstoneRootKey(), domain, project, resourceType, resourceID}, split)
}

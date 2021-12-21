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

package datasource

type TaskFindOptions struct {
	Domain   string
	Project  string
	Action   string
	Status   string
	DataType string
}

type TombstoneFindOptions struct {
	Domain          string
	Project         string
	ResourceType    string
	BeforeTimestamp int64
}

type TaskFindOption func(options *TaskFindOptions)

type TombstoneFindOption func(options *TombstoneFindOptions)

func NewTaskFindOptions() TaskFindOptions {
	return TaskFindOptions{}
}

func NewTombstoneFindOptions() TombstoneFindOptions {
	return TombstoneFindOptions{}
}

// WithDomain find task with domain
func WithDomain(domain string) TaskFindOption {
	return func(options *TaskFindOptions) {
		options.Domain = domain
	}
}

// WithProject find task with project
func WithProject(project string) TaskFindOption {
	return func(options *TaskFindOptions) {
		options.Project = project
	}
}

// WithAction find task with action
func WithAction(action string) TaskFindOption {
	return func(options *TaskFindOptions) {
		options.Action = action
	}
}

// WithStatus find task with status
func WithStatus(status string) TaskFindOption {
	return func(options *TaskFindOptions) {
		options.Status = status
	}
}

// WithDataType find task with dataType
func WithDataType(dataType string) TaskFindOption {
	return func(options *TaskFindOptions) {
		options.DataType = dataType
	}
}

// WithTombstoneDomain find tombstone with domain
func WithTombstoneDomain(domain string) TombstoneFindOption {
	return func(options *TombstoneFindOptions) {
		options.Domain = domain
	}
}

// WithTombstoneProject find tombstone with project
func WithTombstoneProject(project string) TombstoneFindOption {
	return func(options *TombstoneFindOptions) {
		options.Project = project
	}
}

// WithResourceType find tombstone with resource type
func WithResourceType(resourceType string) TombstoneFindOption {
	return func(options *TombstoneFindOptions) {
		options.ResourceType = resourceType
	}
}

// WithBeforeTimestamp find tombstone with beforeTimestamp
func WithBeforeTimestamp(timestamp int64) TombstoneFindOption {
	return func(options *TombstoneFindOptions) {
		options.BeforeTimestamp = timestamp
	}
}

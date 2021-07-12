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

package quota

import (
	"github.com/apache/servicecomb-service-center/datasource"
	"strings"
)

type Options struct {
	Domain         string
	Project        string
	ServiceIDNotIn []string
}

type Option func(opts *Options)

func DomainProject(domainProject string) func(opts *Options) {
	return func(opts *Options) {
		index := strings.Index(domainProject, datasource.SPLIT)
		opts.Domain = domainProject[:index]
		opts.Project = domainProject[index+1:]
	}
}

func ServiceIDNotIn(serviceIDs []string) func(opts *Options) {
	return func(opts *Options) {
		opts.ServiceIDNotIn = serviceIDs
	}
}

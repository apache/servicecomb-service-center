// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/scctl/pkg/writer"
)

var (
	clusterTableHeader = []string{"CLUSTER", "ENDPOINTS"}
)

type ClusterRecord struct {
	Name      string
	Endpoints []string
}

func (s *ClusterRecord) EndpointsString() string {
	return util.StringJoin(s.Endpoints, "\n")
}

func (s *ClusterRecord) PrintBody(fmt string) []string {
	return []string{s.Name, s.EndpointsString()}
}

type ClustersPrinter struct {
	Records map[string]*ClusterRecord
	flags   []interface{}
}

func (sp *ClustersPrinter) SetOutputFormat(f string, all bool) {
	sp.Flags(f, all)
}

func (sp *ClustersPrinter) Flags(flags ...interface{}) []interface{} {
	if len(flags) > 0 {
		sp.flags = flags
	}
	return sp.flags
}

func (sp *ClustersPrinter) PrintBody() (slice [][]string) {
	for _, s := range sp.Records {
		slice = append(slice, s.PrintBody(sp.flags[0].(string)))
	}
	return
}

func (sp *ClustersPrinter) PrintTitle() []string {
	return clusterTableHeader
}

func (sp *ClustersPrinter) Sorter() *writer.RecordsSorter {
	return nil
}

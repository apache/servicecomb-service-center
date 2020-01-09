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

package writer

func firstColumnCmpFunc(row1, row2 []string) bool {
	return row1[0] < row2[0]
}

type RecordsSorter struct {
	Records     [][]string
	CompareFunc func(row1, row2 []string) bool
}

func (s *RecordsSorter) Len() int {
	return len(s.Records)
}

func (s *RecordsSorter) Swap(i, j int) {
	s.Records[i], s.Records[j] = s.Records[j], s.Records[i]
}

func (s *RecordsSorter) Less(i, j int) bool {
	return s.CompareFunc(s.Records[i], s.Records[j])
}

func NewRecordsSorter(cmp func([]string, []string) bool) *RecordsSorter {
	if cmp == nil {
		cmp = firstColumnCmpFunc
	}
	return &RecordsSorter{
		CompareFunc: cmp,
	}
}

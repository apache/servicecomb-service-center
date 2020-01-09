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

import (
	"github.com/olekukonko/tablewriter"
	"os"
	"sort"
	"strconv"
	"time"
)

const Day = time.Hour * 24

type Printer interface {
	Flags(flags ...interface{}) []interface{}
	PrintBody() [][]string
	PrintTitle() []string
	Sorter() *RecordsSorter
}

func TimeFormat(delta time.Duration) string {
	switch {
	case delta < time.Minute:
		return strconv.FormatFloat(delta.Seconds(), 'f', 0, 64) + "s"
	case delta < time.Hour:
		return strconv.FormatFloat(delta.Minutes(), 'f', 0, 64) + "m"
	case delta < Day:
		return strconv.FormatFloat(delta.Hours(), 'f', 0, 64) + "h"
	default:
		return strconv.FormatFloat(float64(delta/Day), 'f', 0, 64) + "d"
	}
}

func Reshape(maxWidth int, line []string) []string {
	for i, col := range line {
		if len(col)-maxWidth > 3 {
			line[i] = col[:maxWidth] + "..."
		}
	}
	return line
}

func MakeTable(tableName []string, tableContent [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(tableName)
	table.SetBorder(false)
	for _, v := range tableContent {
		table.Append(v)
	}
	table.Render()
}

func PrintTable(p Printer) {
	body := p.PrintBody()
	sorter := p.Sorter()
	if sorter == nil {
		sorter = NewRecordsSorter(nil)
	}
	sorter.Records = body
	sort.Sort(sorter)
	MakeTable(p.PrintTitle(), body)
}

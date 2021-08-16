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

package util

import (
	"sort"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/pkg/validate"
)

type VersionRule func(sorted []string, kvs []*kvstore.KeyValue, start, end string) []string

func Sort(kvs []*kvstore.KeyValue, cmp func(start, end string) bool) {
	sorter := newSorter(kvs, cmp, true)
	sort.Sort(sorter)
}

func newSorter(kvs []*kvstore.KeyValue, cmp func(start string, end string) bool, ref bool) *serviceKeySorter {
	tmp := kvs
	if !ref {
		tmp = make([]*kvstore.KeyValue, len(kvs))
	}
	sorter := &serviceKeySorter{
		sortArr: make([]string, len(kvs)),
		kvs:     tmp,
		cmp:     cmp,
	}
	for i, kv := range kvs {
		key := util.BytesToStringWithNoCopy(kv.Key)
		ver := key[strings.LastIndex(key, "/")+1:]
		sorter.sortArr[i] = ver
		sorter.kvs[i] = kv
	}
	return sorter
}

func (vr VersionRule) Match(kvs []*kvstore.KeyValue, ops ...string) []string {
	sorter := newSorter(kvs, Larger, false)
	sort.Sort(sorter)

	args := [2]string{}
	switch {
	case len(ops) > 1:
		args[1] = ops[1]
		fallthrough
	case len(ops) > 0:
		args[0] = ops[0]
	}
	return vr(sorter.sortArr, sorter.kvs, args[0], args[1])
}

type serviceKeySorter struct {
	sortArr []string
	kvs     []*kvstore.KeyValue
	cmp     func(i, j string) bool
}

func (sks *serviceKeySorter) Len() int {
	return len(sks.sortArr)
}

func (sks *serviceKeySorter) Swap(i, j int) {
	sks.sortArr[i], sks.sortArr[j] = sks.sortArr[j], sks.sortArr[i]
	sks.kvs[i], sks.kvs[j] = sks.kvs[j], sks.kvs[i]
}

func (sks *serviceKeySorter) Less(i, j int) bool {
	return sks.cmp(sks.sortArr[i], sks.sortArr[j])
}

func Larger(start, end string) bool {
	s, _ := validate.VersionToInt64(start)
	e, _ := validate.VersionToInt64(end)
	return s > e
}

func LessEqual(start, end string) bool {
	return !Larger(start, end)
}

// Latest return latest version kv
func Latest(sorted []string, kvs []*kvstore.KeyValue, start, end string) []string {
	if len(sorted) == 0 {
		return []string{}
	}
	return []string{kvs[0].Value.(string)}
}

// Range return start <= version < end
func Range(sorted []string, kvs []*kvstore.KeyValue, start, end string) []string {
	total := len(sorted)
	if total == 0 {
		return []string{}
	}

	result := make([]string, 0, total)
	firstFound := false

	if Larger(start, end) {
		start, end = end, start
	}

	eldest, latest := sorted[total-1], sorted[0]
	if Larger(start, latest) || LessEqual(end, eldest) {
		return []string{}
	}

	for i, k := range sorted {
		if !firstFound {
			if LessEqual(end, k) {
				continue
			}
			firstFound = true
		} else if Larger(start, k) {
			break
		}
		// end >= k >= start
		result = append(result, kvs[i].Value.(string))
	}
	return result
}

// AtLess return version >= start
func AtLess(sorted []string, kvs []*kvstore.KeyValue, start, end string) []string {
	total := len(sorted)
	if total == 0 {
		return []string{}
	}

	result := make([]string, 0, total)
	if Larger(start, sorted[0]) {
		return []string{}
	}

	for i, k := range sorted {
		if Larger(start, k) {
			return result[:i]
		}
		result = append(result, kvs[i].Value.(string))
	}
	return result
}

func ParseVersionRule(versionRule string) func(kvs []*kvstore.KeyValue) []string {
	if len(versionRule) == 0 {
		return nil
	}

	rangeIdx := strings.Index(versionRule, "-")
	switch {
	case versionRule == "latest":
		return func(kvs []*kvstore.KeyValue) []string {
			return VersionRule(Latest).Match(kvs)
		}
	case versionRule[len(versionRule)-1:] == "+":
		// 取最低版本及高版本集合
		start := versionRule[:len(versionRule)-1]
		return func(kvs []*kvstore.KeyValue) []string {
			return VersionRule(AtLess).Match(kvs, start)
		}
	case rangeIdx > 0:
		// 取版本范围集合
		start := versionRule[:rangeIdx]
		end := versionRule[rangeIdx+1:]
		return func(kvs []*kvstore.KeyValue) []string {
			return VersionRule(Range).Match(kvs, start, end)
		}
	default:
		// 精确匹配
		return nil
	}
}

func VersionMatchRule(version string, versionRule string) bool {
	match := ParseVersionRule(versionRule)
	if match == nil {
		return version == versionRule
	}

	return len(match([]*kvstore.KeyValue{
		{
			Key:   util.StringToBytesWithNoCopy("/" + version),
			Value: "",
		},
	})) > 0
}

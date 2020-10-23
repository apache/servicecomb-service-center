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
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type VersionRule func(sorted []string, kvs map[string]*sd.KeyValue, start, end string) []string

func (vr VersionRule) Match(kvs []*sd.KeyValue, ops ...string) []string {
	sorter := &serviceKeySorter{
		sortArr: make([]string, len(kvs)),
		kvs:     make(map[string]*sd.KeyValue, len(kvs)),
		cmp:     Larger,
	}
	for i, kv := range kvs {
		key := util.BytesToStringWithNoCopy(kv.Key)
		ver := key[strings.LastIndex(key, "/")+1:]
		sorter.sortArr[i] = ver
		sorter.kvs[ver] = kv
	}
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
	kvs     map[string]*sd.KeyValue
	cmp     func(i, j string) bool
}

func (sks *serviceKeySorter) Len() int {
	return len(sks.sortArr)
}

func (sks *serviceKeySorter) Swap(i, j int) {
	sks.sortArr[i], sks.sortArr[j] = sks.sortArr[j], sks.sortArr[i]
}

func (sks *serviceKeySorter) Less(i, j int) bool {
	return sks.cmp(sks.sortArr[i], sks.sortArr[j])
}

func VersionToInt64(versionStr string) (ret int64, err error) {
	verBytes := [4]int16{}
	idx := 0
	for i := 0; i < 4 && idx < len(versionStr); i++ {
		f := strings.IndexRune(versionStr[idx:], '.')
		if f < 0 {
			f = len(versionStr) - idx
		}
		integer, err := strconv.ParseInt(versionStr[idx:idx+f], 10, 16)
		if err != nil {
			return 0, err
		}
		verBytes[i] = int16(integer)
		idx += f + 1
	}
	ret = util.Int16ToInt64(verBytes[:])
	return
}

func Larger(start, end string) bool {
	s, _ := VersionToInt64(start)
	e, _ := VersionToInt64(end)
	return s > e
}

func LessEqual(start, end string) bool {
	return !Larger(start, end)
}

func Latest(sorted []string, kvs map[string]*sd.KeyValue, start, end string) []string {
	if len(sorted) == 0 {
		return []string{}
	}
	return []string{kvs[sorted[0]].Value.(string)}
}

func Range(sorted []string, kvs map[string]*sd.KeyValue, start, end string) []string {
	result := make([]string, len(sorted))
	i, flag := 0, 0

	if Larger(start, end) {
		start, end = end, start
	}

	l := len(sorted)
	if l == 0 || Larger(start, sorted[0]) || LessEqual(end, sorted[l-1]) {
		return []string{}
	}

	for _, k := range sorted {
		// end >= k >= start
		switch flag {
		case 0:
			if LessEqual(end, k) {
				continue
			}
			flag = 1
		case 1:
			if Larger(start, k) {
				return result[:i]
			}
		}

		result[i] = kvs[k].Value.(string)
		i++
	}
	return result[:i]
}

func AtLess(sorted []string, kvs map[string]*sd.KeyValue, start, end string) []string {
	result := make([]string, len(sorted))

	if len(sorted) == 0 || Larger(start, sorted[0]) {
		return []string{}
	}

	for i, k := range sorted {
		if Larger(start, k) {
			return result[:i]
		}
		result[i] = kvs[k].Value.(string)
	}
	return result[:]
}

func ParseVersionRule(versionRule string) func(kvs []*sd.KeyValue) []string {
	if len(versionRule) == 0 {
		return nil
	}

	rangeIdx := strings.Index(versionRule, "-")
	switch {
	case versionRule == "latest":
		return func(kvs []*sd.KeyValue) []string {
			return VersionRule(Latest).Match(kvs)
		}
	case versionRule[len(versionRule)-1:] == "+":
		// 取最低版本及高版本集合
		start := versionRule[:len(versionRule)-1]
		return func(kvs []*sd.KeyValue) []string {
			return VersionRule(AtLess).Match(kvs, start)
		}
	case rangeIdx > 0:
		// 取版本范围集合
		start := versionRule[:rangeIdx]
		end := versionRule[rangeIdx+1:]
		return func(kvs []*sd.KeyValue) []string {
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

	return len(match([]*sd.KeyValue{
		{
			Key:   util.StringToBytesWithNoCopy("/" + version),
			Value: "",
		},
	})) > 0
}

type VersionRegexp struct {
	Regex *regexp.Regexp
	Fuzzy bool
}

func (vr *VersionRegexp) MatchString(s string) bool {
	if vr.Regex != nil && !vr.Regex.MatchString(s) {
		return false
	}
	return vr.validateVersionRule(s) == nil
}

func (vr *VersionRegexp) String() string {
	if vr.Fuzzy {
		return "the form x[.y[.z]] or x[.y[.z]]+ or x[.y[.z]]-x[.y[.z]] or 'latest' where x y and z are 0-32767 range"
	}
	return "the form x[.y[.z]] where x y and z are 0-32767 range"
}

func (vr *VersionRegexp) validateVersionRule(versionRule string) (err error) {
	if len(versionRule) == 0 {
		return
	}

	if !vr.Fuzzy {
		_, err = VersionToInt64(versionRule)
		return
	}

	rangeIdx := strings.Index(versionRule, "-")
	switch {
	case versionRule == "latest":
		return
	case versionRule[len(versionRule)-1:] == "+":
		// 取最低版本及高版本集合
		start := versionRule[:len(versionRule)-1]
		_, err = VersionToInt64(start)
	case rangeIdx > 0:
		// 取版本范围集合
		start := versionRule[:rangeIdx]
		end := versionRule[rangeIdx+1:]
		_, err = VersionToInt64(start)
		if err == nil {
			_, err = VersionToInt64(end)
		}
	default:
		// 精确匹配
		_, err = VersionToInt64(versionRule)
	}
	return
}

func NewVersionRegexp(fuzzy bool) (vr *VersionRegexp) {
	vr = &VersionRegexp{Fuzzy: fuzzy}
	if fuzzy {
		vr.Regex, _ = regexp.Compile(`^\d+(\.\d+){0,3}\+?$|^\d+(\.\d+){0,3}-\d+(\.\d+){0,3}$|^latest$`)
		return
	}
	vr.Regex, _ = regexp.Compile(`^\d+(\.\d+){0,3}$`)
	return
}

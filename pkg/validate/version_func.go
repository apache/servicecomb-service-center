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

package validate

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/util"
)

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
		if _, err = VersionToInt64(start); err == nil {
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

func VersionToInt64(versionStr string) (int64, error) {
	verBytes := [4]int16{}

	for idx, i := 0, 0; i < 4 && idx < len(versionStr); i++ {
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
	ret := util.Int16ToInt64(verBytes[:])
	return ret, nil
}

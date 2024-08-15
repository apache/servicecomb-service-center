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
	"bytes"
	"runtime"
	"strings"
	"unsafe"

	"github.com/cloudflare/gokey"
)

const TypePass = "pass"

var passwordSpec = &gokey.PasswordSpec{
	Length:         8,
	Upper:          1,
	Lower:          1,
	Digits:         1,
	Special:        1,
	AllowedSpecial: "-~!@#$%^&*()_=+|<>{}[]",
}

func SafeCloseChan(c chan struct{}) {
	if c == nil {
		return
	}

	select {
	case _, ok := <-c:
		if ok {
			close(c)
		}
	default:
		close(c)
	}
}

func BytesToStringWithNoCopy(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

func StringToBytesWithNoCopy(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

func ListToMap(list []string) map[string]struct{} {
	ret := make(map[string]struct{}, len(list))
	for _, v := range list {
		ret[v] = struct{}{}
	}
	return ret
}

func MapToList(dict map[string]struct{}) []string {
	ret := make([]string, 0, len(dict))
	for k := range dict {
		ret = append(ret, k)
	}
	return ret
}

func StringJoin(args []string, sep string) string {
	l := len(args)
	switch l {
	case 0:
		return ""
	case 1:
		return args[0]
	case 2:
		return args[0] + sep + args[1]
	default:
		n := len(sep) * (l - 1)
		for i := 0; i < l; i++ {
			n += len(args[i])
		}
		b := make([]byte, n)
		sl := copy(b, args[0])
		for i := 1; i < l; i++ {
			sl += copy(b[sl:], sep)
			sl += copy(b[sl:], args[i])
		}
		return BytesToStringWithNoCopy(b)
	}
}

func GetCaller(skip int) (string, string, int, bool) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	method := FormatFuncName(runtime.FuncForPC(pc).Name())
	return file, method, line, ok
}

func Int16ToInt64(bs []int16) (in int64) {
	l := len(bs)
	if l > 4 || l == 0 {
		return 0
	}

	pi := (*[4]int16)(unsafe.Pointer(&in))
	if IsBigEndian() {
		for i := range bs {
			pi[i] = bs[l-i-1]
		}
		return
	}

	for i := range bs {
		pi[3-i] = bs[l-i-1]
	}
	return
}

func SliceHave(arr []string, str string) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}

func StringTRUE(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "1" || s == "true" {
		return true
	}
	return false
}

func FromDomainProject(domainProject string) (domain, project string) {
	if i := strings.Index(domainProject, "/"); i >= 0 {
		return domainProject[:i], domainProject[i+1:]
	}
	return domainProject, ""
}

func ToDomainProject(domain, project string) (domainProject string) {
	domainProject = domain + "/" + project
	return
}

func IsVersionOrHealthPattern(pattern string) bool {
	return strings.HasSuffix(pattern, "/version") || strings.HasSuffix(pattern, "/health")
}

func ToSnake(name string) string {
	if name == "" {
		return ""
	}
	temp := strings.Split(name, "-")
	var buffer bytes.Buffer
	for num, v := range temp {
		vv := []rune(v)
		if num == 0 {
			buffer.WriteString(string(vv))
			continue
		}
		if len(vv) > 0 {
			if vv[0] >= 'a' && vv[0] <= 'z' { //首字母大写
				vv[0] -= 32
			}
			buffer.WriteString(string(vv))
		}
	}
	return buffer.String()
}

func GeneratePassword() (string, error) {
	seed, err := gokey.GenerateEncryptedKeySeed(TypePass)
	if err != nil {
		return "", err
	}
	pass, err := gokey.GetPass(TypePass, "", seed, passwordSpec)
	if err != nil {
		return "", err
	}
	return pass, nil
}

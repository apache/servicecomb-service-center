//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package util

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unsafe"
)

const (
	INIT_FAIL_EXIT = 2
)

func GetAppPath(path string) string {
	env := os.Getenv("APP_ROOT")
	if len(env) == 0 {
		wd, _ := os.Getwd()
		return filepath.Join(wd, path)
	}
	return os.ExpandEnv(filepath.Join("$APP_ROOT", path))
}

func PathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func MinInt(x, y int) int {
	if x <= y {
		return x
	} else {
		return y
	}
}

func ClearStringMemory(src *string) {
	p := (*struct {
		ptr uintptr
		len int
	})(unsafe.Pointer(src))

	len := MinInt(p.len, 32)
	ptr := p.ptr
	for idx := 0; idx < len; idx = idx + 1 {
		b := (*byte)(unsafe.Pointer(ptr))
		*b = 0
		ptr += 1
	}
}

func ClearByteMemory(src []byte) {
	len := MinInt(len(src), 32)
	for idx := 0; idx < len; idx = idx + 1 {
		src[idx] = 0
	}
}

func ParaseTenantProject(ctx context.Context) string {
	tenant := ctx.Value("tenant").(string)
	project := ctx.Value("project").(string)
	tenant = strings.Join([]string{tenant, project}, "/")
	return tenant
}

func ParaseTenant(ctx context.Context) string {
	return ctx.Value("tenant").(string)
}

//format : https://10.21.119.167:30100 or http://10.21.119.167:30100
func URLChecker(url string) (bool, error) {
	ipPatten := "((?:(?:25[0-5]|2[0-4]\\d|((1\\d{2})|([1-9]?\\d)))\\.){3}(?:25[0-5]|2[0-4]\\d|((1\\d{2})|([1-9]?\\d))))"
	patten := "^(https|http):\\/\\/" + ipPatten + ":([0-9]+)$"
	ok, err := regexp.MatchString(patten, url)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func MapChecker(data map[string]string) bool {
	if data == nil {
		return false
	}
	if len(data) == 0 {
		return false
	}
	for key, value := range data {
		if len(key) == 0 {
			return false
		}
		if len(value) == 0 {
			return false
		}
	}
	return true
}

func GetIPFromContext(ctx context.Context) string {
	remoteIp := ""
	remoteIp, _ = ctx.Value("x-remote-ip").(string)
	return remoteIp
}

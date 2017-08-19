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
	"bytes"
	"encoding/gob"
	"golang.org/x/net/context"
	"os"
	"path/filepath"
	"regexp"
	"time"
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

type StringContext struct {
	parentCtx context.Context
	kv        map[string]interface{}
}

func (c *StringContext) Deadline() (deadline time.Time, ok bool) {
	return c.parentCtx.Deadline()
}

func (c *StringContext) Done() <-chan struct{} {
	return c.parentCtx.Done()
}

func (c *StringContext) Err() error {
	return c.parentCtx.Err()
}

func (c *StringContext) Value(key interface{}) interface{} {
	k, ok := key.(string)
	if !ok {
		return c.parentCtx.Value(key)
	}
	return c.kv[k]
}

func (c *StringContext) SetKV(key string, val interface{}) {
	c.kv[key] = val
}

func NewContext(ctx context.Context, key string, val interface{}) context.Context {
	strCtx, ok := ctx.(*StringContext)
	if !ok {
		strCtx = &StringContext{
			parentCtx: ctx,
			kv:        make(map[string]interface{}, 10),
		}
	}
	strCtx.SetKV(key, val)
	return strCtx
}

func FromContext(ctx context.Context, key string) interface{} {
	return ctx.Value(key)
}

func ParseTenantProject(ctx context.Context) string {
	return StringJoin([]string{ParseTenant(ctx), ParseProject(ctx)}, "/")
}

func ParseTenant(ctx context.Context) string {
	v, ok := FromContext(ctx, "tenant").(string)
	if !ok {
		return ""
	}
	return v
}

func ParseProject(ctx context.Context) string {
	v, ok := FromContext(ctx, "project").(string)
	if !ok {
		return ""
	}
	return v
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

func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func SafeCloseChan(c chan struct{}) {
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
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
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

func RecoverAndReport() {
	if r := recover(); r != nil {
		LOGGER.Errorf(nil, "recover! %v", r)
	}
}

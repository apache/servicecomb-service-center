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
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInt16ToInt64(t *testing.T) {
	bs := []int16{0, 0, 0, 1}
	i := Int16ToInt64(bs)
	if i != 1 {
		t.Fatalf("Int16ToInt64 failed, %v %d != %d", bs, i, 1)
	}

	bs = []int16{1, 0, 0, 0}
	i = Int16ToInt64(bs)
	if i != 1<<(3*16) {
		t.Fatalf("Int16ToInt64 failed, %v %d != %d", bs, i, 1<<(3*16))
	}

	bs = []int16{1}
	i = Int16ToInt64(bs)
	if i != 1 {
		t.Fatalf("Int16ToInt64 failed, %v %d != %d", bs, i, 1)
	}

	bs = []int16{1, 0}
	i = Int16ToInt64(bs)
	if i != 1<<16 {
		t.Fatalf("Int16ToInt64 failed, %v %d != %d", bs, i, 1<<16)
	}
}

func TestStringJoin(t *testing.T) {
	if StringJoin([]string{"a", "b", "c"}, ",") != "a,b,c" {
		t.Fatalf("TestStringJoin failed")
	}
	if StringJoin([]string{"a"}, ",") != "a" {
		t.Fatalf("TestStringJoin failed")
	}
	if StringJoin([]string{"a", "b", "c"}, "") != "abc" {
		t.Fatalf("TestStringJoin failed")
	}
	if StringJoin([]string{}, ",") != "" {
		t.Fatalf("TestStringJoin failed")
	}
	if StringJoin(nil, ",") != "" {
		t.Fatalf("TestStringJoin failed")
	}
}

func TestStringToBytesWithNoCopy(t *testing.T) {
	b := StringToBytesWithNoCopy("ab")
	if b[0] != 'a' || b[1] != 'b' {
		t.Fatalf("TestStringToBytesWithNoCopy failed")
	}
}

func TestListToMap(t *testing.T) {
	m := ListToMap(nil)
	if m == nil || len(m) > 0 {
		t.Fatalf("TestListToMap falied")
	}
	m = ListToMap([]string{})
	if m == nil || len(m) > 0 {
		t.Fatalf("TestListToMap falied")
	}
	m = ListToMap([]string{"a"})
	if m == nil || len(m) != 1 || m["a"] != struct{}{} {
		t.Fatalf("TestListToMap falied")
	}
}

func TestSafeCloseChan(t *testing.T) {
	var ch chan struct{}
	SafeCloseChan(ch)
	ch = make(chan struct{})
	SafeCloseChan(ch)
	SafeCloseChan(ch)
}

func TestSystemPackage(t *testing.T) {
	if HostName() == "" {
		t.Fatalf("TestSystemPackage failed")
	}
	if !PathExist("../../etc/conf/app.conf") {
		t.Fatalf("TestSystemPackage failed")
	}
}

func TestGetEnvInt(t *testing.T) {
	os.Unsetenv("a")
	if GetEnvInt("a", 1) != 1 {
		t.Fatalf("TestGetEnvInt failed")
	}
	os.Setenv("a", "")
	if GetEnvInt("a", 1) != 1 {
		t.Fatalf("TestGetEnvInt failed")
	}
	os.Setenv("a", "x")
	if GetEnvInt("a", 1) != 1 {
		t.Fatalf("TestGetEnvInt failed")
	}
	os.Setenv("a", "2")
	if GetEnvInt("a", 1) != 2 {
		t.Fatalf("TestGetEnvInt failed")
	}
}

func TestGetEnvString(t *testing.T) {
	os.Unsetenv("a")
	if GetEnvString("a", "1") != "1" {
		t.Fatalf("TestGetEnvInt failed")
	}
	os.Setenv("a", "")
	if GetEnvString("a", "1") != "" {
		t.Fatalf("TestGetEnvInt failed")
	}
	os.Setenv("a", "2")
	if GetEnvString("a", "1") != "2" {
		t.Fatalf("TestGetEnvInt failed")
	}
}

func TestBytesToStringWithNoCopy(t *testing.T) {
	s := BytesToStringWithNoCopy(nil)
	if s != "" {
		t.Fatal("TestBytesToStringWithNoCopy failed")
	}
}

func TestIsVersionOrHealthPattern(t *testing.T) {
	assert.True(t, IsVersionOrHealthPattern("/version"))
	assert.True(t, IsVersionOrHealthPattern("/v4/a/registry/version"))
	assert.False(t, IsVersionOrHealthPattern("/version/a"))
	assert.True(t, IsVersionOrHealthPattern("/health"))
	assert.True(t, IsVersionOrHealthPattern("/v4/a/registry/health"))
	assert.False(t, IsVersionOrHealthPattern("/health/a"))
}

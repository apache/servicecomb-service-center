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
	"testing"
)

type S struct {
	A int
	B string
	C bool
	D [1]string
	E []string
	F map[string]string
	G C
	H *C
	I [1]C
	J []C
	K [1]*C
	L []*C
	M map[string]C
	N map[string]*C
	O *CC
}

type C struct {
	B string
}

type CC struct {
	CH *C
}

func TestValidator_Validate(t *testing.T) {
	v := Validator{}

	s := new(S)

	v.AddRule("A", &ValidateRule{Min: 1, Max: 2})
	err := v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.A = 3
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.A = 1
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.A = 2
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	r, _ := regexp.Compile(`^\d*$`)
	v.AddRule("B", &ValidateRule{Min: 1, Max: 2, Regexp: r})
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.B = "111"
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.B = "2"
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.B = "33"
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	v.AddRule("D", &ValidateRule{Regexp: r})
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.D = [1]string{"aaa"}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.D = [1]string{"1"}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	v.AddRule("E", &ValidateRule{Min: 1, Max: 2, Regexp: r})
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.E = []string{"1", "2", "3"}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	s.E = []string{"1", "aaa"}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.E = []string{"1"}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	v.AddRule("F", &ValidateRule{Min: 1, Max: 2, Regexp: r})
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.F = map[string]string{"1": "", "2": "", "3": ""}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	s.F = map[string]string{"1": "", "aaa": ""}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.F = map[string]string{"1": "", "2": "aaa"}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.F = map[string]string{"1": ""}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	// sub struct
	pr, _ := regexp.Compile(`^[a-z]*$`)
	v.AddRule("G", &ValidateRule{Regexp: pr})
	v.AddSub("G", &v)
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.G = C{B: "111"}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.G = C{B: "2"}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.G = C{B: "33"}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	// just check sub
	var hv Validator
	hv.AddRule("H", &ValidateRule{Min: 1, Max: 2, Regexp: pr})
	v.AddSub("H", &hv)
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	// also check nil pointer
	v.AddRule("H", &ValidateRule{Min: 1, Max: 2, Regexp: pr})
	v.AddSub("H", &v)
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.H = &C{B: "111"}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.H = &C{B: "a"}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.H = &C{B: "2"}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.H = &C{B: "33"}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	v.AddRule("I", &ValidateRule{Regexp: pr})
	v.AddSub("I", &v)
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.I = [1]C{{B: "111"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.I = [1]C{{B: "a"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.I = [1]C{{B: "2"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.I = [1]C{{B: "33"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	v.AddRule("J", &ValidateRule{Min: 1, Max: 2, Regexp: pr})
	v.AddSub("J", &v)
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.J = []C{{B: "111"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.J = []C{{B: "a"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.J = []C{{B: "2"}, {B: "2"}, {B: "2"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.J = []C{{B: "2"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.J = []C{{B: "33"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	v.AddRule("K", &ValidateRule{Min: 1, Regexp: pr})
	v.AddSub("K", &v)
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.K = [1]*C{{B: "111"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.K = [1]*C{{B: "a"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.K = [1]*C{{B: "2"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.K = [1]*C{{B: "33"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	v.AddRule("L", &ValidateRule{Min: 1, Max: 2, Regexp: pr})
	v.AddSub("L", &v)
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.L = []*C{{B: "111"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.L = []*C{{B: "a"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.L = []*C{{B: "2"}, {B: "2"}, {B: "2"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.L = []*C{{B: "2"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.L = []*C{{B: "33"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	v.AddRule("M", &ValidateRule{Min: 1, Max: 2, Regexp: pr})
	v.AddSub("M", &v)
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.M = map[string]C{"a": {B: "111"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.M = map[string]C{"a": {B: "a"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.M = map[string]C{"a": {B: "1"}, "b": {B: "1"}, "c": {B: "1"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.M = map[string]C{"a": {B: "2"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.M = map[string]C{"a": {B: "33"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	v.AddRule("N", &ValidateRule{Min: 1, Max: 2, Regexp: pr})
	v.AddSub("N", &v)
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.N = map[string]*C{"a": {B: "111"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.N = map[string]*C{"a": {B: "a"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.N = map[string]*C{"a": {B: "1"}, "b": {B: "1"}, "c": {B: "1"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.N = map[string]*C{"a": {B: "2"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.N = map[string]*C{"a": {B: "33"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}

	v.AddRule("O", &ValidateRule{Min: 1, Max: 2, Regexp: pr})
	v.AddSub("O", &v)
	v.AddRule("CH", &ValidateRule{Min: 1, Max: 2, Regexp: pr})
	v.AddSub("CH", &v)
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.O = &CC{CH: &C{B: "111"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.O = &CC{CH: &C{B: "a"}}
	err = v.Validate(s)
	if err == nil {
		t.Fatalf("validate failed")
	}
	println(err.Error())
	s.O = &CC{CH: &C{B: "2"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
	s.O = &CC{CH: &C{B: "33"}}
	err = v.Validate(s)
	if err != nil {
		t.Fatalf("validate failed, %s", err.Error())
	}
}

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
	"fmt"
	"reflect"
	"sort"
	"testing"

	sd "github.com/apache/servicecomb-service-center/datasource/etcd/pkg/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/stretchr/testify/assert"
)

const VERSIONRULE_BASE = 5000

func BenchmarkVersionRule_Latest_GetServicesIds(b *testing.B) {
	var kvs = make([]*sd.KeyValue, VERSIONRULE_BASE)
	for i := 1; i <= VERSIONRULE_BASE; i++ {
		kvs[i-1] = &sd.KeyValue{
			Key:   []byte(fmt.Sprintf("/service/ver/1.%d", i)),
			Value: []byte(fmt.Sprintf("%d", i)),
		}
	}
	b.N = VERSIONRULE_BASE
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VersionRule(Latest).Match(kvs)
	}
	b.ReportAllocs()
	// 5000	   7105020 ns/op	 2180198 B/op	   39068 allocs/op
}

func BenchmarkVersionRule_Range_GetServicesIds(b *testing.B) {
	var kvs = make([]*sd.KeyValue, VERSIONRULE_BASE)
	for i := 1; i <= VERSIONRULE_BASE; i++ {
		kvs[i-1] = &sd.KeyValue{
			Key:   []byte(fmt.Sprintf("/service/ver/1.%d", i)),
			Value: []byte(fmt.Sprintf("%d", i)),
		}
	}
	b.N = VERSIONRULE_BASE
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VersionRule(Range).Match(kvs, fmt.Sprintf("1.%d", i), fmt.Sprintf("1.%d", i+VERSIONRULE_BASE/10))
	}
	b.ReportAllocs()
	// 5000	   7244029 ns/op	 2287389 B/op	   39584 allocs/op
}

func BenchmarkVersionRule_AtLess_GetServicesIds(b *testing.B) {
	var kvs = make([]*sd.KeyValue, VERSIONRULE_BASE)
	for i := 1; i <= VERSIONRULE_BASE; i++ {
		kvs[i-1] = &sd.KeyValue{
			Key:   []byte(fmt.Sprintf("/service/ver/1.%d", i)),
			Value: []byte(fmt.Sprintf("%d", i)),
		}
	}
	b.N = VERSIONRULE_BASE
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VersionRule(AtLess).Match(kvs, fmt.Sprintf("1.%d", i))
	}
	b.ReportAllocs()
	// 5000	  11221098 ns/op	 3174720 B/op	   58064 allocs/op
}

func BenchmarkParseVersionRule(b *testing.B) {
	f := ParseVersionRule("latest")
	kvs := []*sd.KeyValue{
		{
			Key:   []byte("/service/ver/1.0.300"),
			Value: "1.0.300",
		},
		{
			Key:   []byte("/service/ver/1.0.303"),
			Value: "1.0.303",
		},
		{
			Key:   []byte("/service/ver/1.0.304"),
			Value: "1.0.304",
		},
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(kvs)
		}
	})
	b.ReportAllocs()
}

func TestSorter(t *testing.T) {
	log.Info("normal")

	t.Run("version asc", func(t *testing.T) {
		kvs := []string{"1.0.0", "1.0.1"}
		sort.Sort(&serviceKeySorter{
			sortArr: kvs,
			kvs:     make(map[string]*sd.KeyValue),
			cmp:     Larger,
		})
		assert.Equal(t, "1.0.1", kvs[0])
		assert.Equal(t, "1.0.0", kvs[1])
	})

	t.Run("version desc", func(t *testing.T) {
		kvs := []string{"1.0.1", "1.0.0"}
		sort.Sort(&serviceKeySorter{
			sortArr: kvs,
			kvs:     make(map[string]*sd.KeyValue),
			cmp:     Larger,
		})
		assert.Equal(t, "1.0.1", kvs[0])
		assert.Equal(t, "1.0.0", kvs[1])
	})

	t.Run("len(v1) != len(v2)", func(t *testing.T) {
		kvs := []string{"1.0.0.0", "1.0.1"}
		sort.Sort(&serviceKeySorter{
			sortArr: kvs,
			kvs:     make(map[string]*sd.KeyValue),
			cmp:     Larger,
		})
		assert.Equal(t, "1.0.1", kvs[0])
		assert.Equal(t, "1.0.0.0", kvs[1])
	})

	t.Run("1.0.9 vs 1.0.10", func(t *testing.T) {
		kvs := []string{"1.0.9", "1.0.10"}
		sort.Sort(&serviceKeySorter{
			sortArr: kvs,
			kvs:     make(map[string]*sd.KeyValue),
			cmp:     Larger,
		})
		assert.Equal(t, "1.0.10", kvs[0])
		assert.Equal(t, "1.0.9", kvs[1])
	})

	t.Run("1.10 vs 4", func(t *testing.T) {
		kvs := []string{"1.10", "4"}
		sort.Sort(&serviceKeySorter{
			sortArr: kvs,
			kvs:     make(map[string]*sd.KeyValue),
			cmp:     Larger,
		})
		assert.Equal(t, "4", kvs[0])
		assert.Equal(t, "1.10", kvs[1])
	})

	log.Info("exception")

	t.Run("invalid version1", func(t *testing.T) {
		kvs := []string{"1.a", "1.0.1.a", ""}
		sort.Sort(&serviceKeySorter{
			sortArr: kvs,
			kvs:     make(map[string]*sd.KeyValue),
			cmp:     Larger,
		})
		assert.Equal(t, "1.a", kvs[0])
		assert.Equal(t, "1.0.1.a", kvs[1])
		assert.Equal(t, "", kvs[2])
	})

	t.Run("invalid version2 > 32767", func(t *testing.T) {
		kvs := []string{"1.0", "1.0.1.32768"}
		sort.Sort(&serviceKeySorter{
			sortArr: kvs,
			kvs:     make(map[string]*sd.KeyValue),
			cmp:     Larger,
		})
		assert.Equal(t, "1.0", kvs[0])
		assert.Equal(t, "1.0.1.32768", kvs[1])
		kvs = []string{"1.0", "1.0.1.32767"}
		sort.Sort(&serviceKeySorter{
			sortArr: kvs,
			kvs:     make(map[string]*sd.KeyValue),
			cmp:     Larger,
		})
		assert.Equal(t, "1.0.1.32767", kvs[0])
		assert.Equal(t, "1.0", kvs[1])
	})
}

func TestVersionRule(t *testing.T) {
	const count = 10
	var kvs = [count]*sd.KeyValue{}
	for i := 1; i <= count; i++ {
		kvs[i-1] = &sd.KeyValue{
			Key:   []byte(fmt.Sprintf("/service/ver/1.%d", i)),
			Value: fmt.Sprintf("%d", i),
		}
	}

	log.Info("normal")

	t.Run("latest", func(t *testing.T) {
		results := VersionRule(Latest).Match(kvs[:])
		assert.Equal(t, 1, len(results))
		assert.Equal(t, fmt.Sprintf("%d", count), results[0])
	})

	t.Run("range1.1 ver in [1.4, 1.8)", func(t *testing.T) {
		results := VersionRule(Range).Match(kvs[:], "1.4", "1.8")
		assert.Equal(t, 4, len(results))
		assert.Equal(t, "7", results[0])
		assert.Equal(t, "4", results[3])
	})

	t.Run("range1.2 ver in (1.8, 1.4]", func(t *testing.T) {
		results := VersionRule(Range).Match(kvs[:], "1.8", "1.4")
		assert.Equal(t, 4, len(results))
		assert.Equal(t, "7", results[0])
		assert.Equal(t, "4", results[3])
	})

	t.Run("range2 ver in [1, 2]", func(t *testing.T) {
		results := VersionRule(Range).Match(kvs[:], "1", "2")
		assert.Equal(t, 10, len(results))
		assert.Equal(t, "10", results[0])
		assert.Equal(t, "1", results[9])
	})

	t.Run("range3 ver in [1.4.1, 1.9.1]", func(t *testing.T) {
		results := VersionRule(Range).Match(kvs[:], "1.4.1", "1.9.1")
		assert.Equal(t, 5, len(results))
		assert.Equal(t, "9", results[0])
		assert.Equal(t, "5", results[4])
	})

	t.Run("range4 ver in [2, 4]", func(t *testing.T) {
		results := VersionRule(Range).Match(kvs[:], "2", "4")
		assert.Equal(t, len(results), 0)
	})

	t.Run("atLess1 ver >= 1.6", func(t *testing.T) {
		results := VersionRule(AtLess).Match(kvs[:], "1.6")
		assert.Equal(t, len(results), 5)
		assert.Equal(t, "10", results[0])
		assert.Equal(t, "6", results[4])
	})

	t.Run("atLess2 ver >= 1", func(t *testing.T) {
		results := VersionRule(AtLess).Match(kvs[:], "1")
		assert.Equal(t, len(results), 10)
		assert.Equal(t, "10", results[0])
		assert.Equal(t, "1", results[9])
	})

	t.Run("atLess3 ver >= 1.5.1", func(t *testing.T) {
		results := VersionRule(AtLess).Match(kvs[:], "1.5.1")
		assert.Equal(t, 5, len(results))
		assert.Equal(t, "10", results[0])
		assert.Equal(t, "6", results[4])
	})

	t.Run("atLess4 ver >= 2", func(t *testing.T) {
		results := VersionRule(AtLess).Match(kvs[:], "2")
		assert.Equal(t, 0, len(results))
	})

	log.Info("exception")

	t.Run("nil", func(t *testing.T) {
		results := VersionRule(Latest).Match(nil)
		assert.Equal(t, 0, len(results))
		results = VersionRule(AtLess).Match(nil)
		assert.Equal(t, 0, len(results))
		results = VersionRule(Range).Match(nil)
		assert.Equal(t, 0, len(results))
		rule := ParseVersionRule("")
		assert.Equal(t, true, reflect.ValueOf(rule).IsNil())
		rule = ParseVersionRule("abc")
		assert.Equal(t, true, reflect.ValueOf(rule).IsNil())
		assert.Equal(t, true, VersionMatchRule("1.0", "1.0"))
		assert.Equal(t, false, VersionMatchRule("1.0", "1.2"))
	})

	log.Info("parse")

	t.Run("latest", func(t *testing.T) {
		match := ParseVersionRule("latest")
		results := match(kvs[:])
		assert.Equal(t, 1, len(results))
		assert.Equal(t, fmt.Sprintf("%d", count), results[0])
	})

	t.Run("range ver in[1.4, 1.8]", func(t *testing.T) {
		match := ParseVersionRule("1.4-1.8")
		results := match(kvs[:])
		assert.Equal(t, 4, len(results))
		assert.Equal(t, "7", results[0])
		assert.Equal(t, "4", results[3])
	})

	t.Run("atLess ver >= 1.6", func(t *testing.T) {
		match := ParseVersionRule("1.6+")
		results := match(kvs[:])
		assert.Equal(t, 5, len(results))
		assert.Equal(t, "10", results[0])
		assert.Equal(t, "6", results[4])
	})

	log.Info("version match rule")

	t.Run("latest", func(t *testing.T) {
		assert.Equal(t, true, VersionMatchRule("1.0", "latest"))
	})

	t.Run("range ver in [1.4, 1.8]", func(t *testing.T) {
		assert.Equal(t, true, VersionMatchRule("1.4", "1.4-1.8"))
		assert.Equal(t, true, VersionMatchRule("1.6", "1.4-1.8"))
		assert.Equal(t, false, VersionMatchRule("1.8", "1.4-1.8"))
		assert.Equal(t, false, VersionMatchRule("1.0", "1.4-1.8"))
		assert.Equal(t, false, VersionMatchRule("1.9", "1.4-1.8"))
	})

	t.Run("atLess ver >= 1.6", func(t *testing.T) {
		assert.Equal(t, true, VersionMatchRule("1.6", "1.6+"))
		assert.Equal(t, true, VersionMatchRule("1.9", "1.6+"))
		assert.Equal(t, false, VersionMatchRule("1.0", "1.6+"))
	})
}

func TestNewVersionRegexp(t *testing.T) {
	log.Info("normal")

	t.Run("latest", func(t *testing.T) {
		vr := NewVersionRegexp(false)
		assert.Equal(t, false, vr.MatchString("latest"))
		vr = NewVersionRegexp(true)
		assert.Equal(t, true, vr.MatchString("latest"))
	})

	t.Run("range", func(t *testing.T) {
		vr := NewVersionRegexp(false)
		assert.Equal(t, false, vr.MatchString("1.1-2.2"))
		vr = NewVersionRegexp(true)
		assert.Equal(t, false, vr.MatchString("-"))
		assert.Equal(t, false, vr.MatchString("1.1-"))
		assert.Equal(t, false, vr.MatchString("-1.1"))
		assert.Equal(t, false, vr.MatchString("1.a-2.b"))
		assert.Equal(t, false, vr.MatchString("1.-.2"))
		assert.Equal(t, false, vr.MatchString("60000-1"))
		assert.Equal(t, true, vr.MatchString("1.1-2.2"))
		assert.Equal(t, true, vr.MatchString("1.1.1.1-2.2.2.2"))
		assert.Equal(t, false, vr.MatchString("1.1.1.1.1-2.2.2.2"))
		assert.Equal(t, false, vr.MatchString("1.1.1.1-2.2.2.2.2"))
	})

	t.Run("atLess", func(t *testing.T) {
		vr := NewVersionRegexp(false)
		assert.Equal(t, false, vr.MatchString("1.0+"))
		vr = NewVersionRegexp(true)
		assert.Equal(t, false, vr.MatchString("+"))
		assert.Equal(t, false, vr.MatchString("+1.0"))
		assert.Equal(t, false, vr.MatchString("1.a+"))
		assert.Equal(t, false, vr.MatchString(".1+"))
		assert.Equal(t, false, vr.MatchString("1.+"))
		assert.Equal(t, false, vr.MatchString(".+"))
		assert.Equal(t, false, vr.MatchString("60000+"))
		assert.Equal(t, true, vr.MatchString("1.0+"))
		assert.Equal(t, true, vr.MatchString("1.0.0.0+"))
		assert.Equal(t, false, vr.MatchString("1.0.0.0.0+"))
	})

	t.Run("explicit", func(t *testing.T) {
		vr := NewVersionRegexp(false)
		assert.Equal(t, false, vr.MatchString(""))
		assert.Equal(t, false, vr.MatchString("a"))
		assert.Equal(t, false, vr.MatchString("60000"))
		assert.Equal(t, false, vr.MatchString("."))
		assert.Equal(t, false, vr.MatchString("1."))
		assert.Equal(t, false, vr.MatchString(".1"))
		assert.Equal(t, true, vr.MatchString("1.4"))
		vr = NewVersionRegexp(true)
		assert.Equal(t, false, vr.MatchString(""))
		assert.Equal(t, false, vr.MatchString("a"))
		assert.Equal(t, false, vr.MatchString("60000"))
		assert.Equal(t, false, vr.MatchString("."))
		assert.Equal(t, false, vr.MatchString("1."))
		assert.Equal(t, false, vr.MatchString(".1"))
		assert.Equal(t, true, vr.MatchString("1.4"))
		assert.Equal(t, true, vr.MatchString("1.4.0.0"))
		assert.Equal(t, false, vr.MatchString("1.4.0.0.0"))
	})

	log.Info("exception")

	t.Run("MatchString & String", func(t *testing.T) {
		vr := VersionRegexp{}
		assert.Equal(t, true, vr.MatchString(""))
		assert.NotEqual(t, "", vr.String())
		vr = VersionRegexp{Fuzzy: true}
		assert.Equal(t, true, vr.MatchString(""))
		assert.NotEqual(t, "", vr.String())
	})
}

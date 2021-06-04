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

package privacy_test

import (
	"crypto/sha512"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/privacy"
	scrypt "github.com/elithrar/simple-scrypt"
	"github.com/go-chassis/foundation/stringutil"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/pbkdf2"
)

func TestHashPassword(t *testing.T) {
	h, _ := privacy.HashPassword("test")
	t.Log(h)
	mac, _ := privacy.ScryptPassword("test")
	t.Log(mac)

	t.Run("given old hash result, should be compatible", func(t *testing.T) {
		same := privacy.SamePassword(h, "test")
		assert.True(t, same)
	})

	sameMac := privacy.SamePassword(mac, "test")
	assert.True(t, sameMac)

	t.Run("use different params for scrypt, should be compatible", func(t *testing.T) {
		h2, _ := scrypt.GenerateFromPassword([]byte("test"), scrypt.Params{N: 1024, R: 8, P: 1, SaltLen: 8, DKLen: 32})
		same := privacy.SamePassword(stringutil.Bytes2str(h2), "test")
		assert.True(t, same)
	})
}
func BenchmarkBcrypt(b *testing.B) {
	h, _ := privacy.HashPassword("test")
	for i := 0; i < b.N; i++ {
		same := privacy.SamePassword(h, "test")
		if !same {
			panic("")
		}

	}
	b.ReportAllocs()
}
func BenchmarkScrypt(b *testing.B) {
	h, _ := privacy.ScryptPassword("test")
	for i := 0; i < b.N; i++ {
		same := privacy.SamePassword(h, "test")
		if !same {
			panic("")
		}

	}
	b.ReportAllocs()
}
func BenchmarkScryptP(b *testing.B) {
	h, _ := privacy.ScryptPassword("test")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			same := privacy.SamePassword(h, "test")
			if !same {
				panic("")
			}
		}
	})
	b.ReportAllocs()
}
func BenchmarkScrypt1024(b *testing.B) {
	p := scrypt.Params{N: 1024, R: 8, P: 1, SaltLen: 8, DKLen: 32}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = scrypt.GenerateFromPassword([]byte("test"), p)
		}
	})
	b.ReportAllocs()
}
func BenchmarkScrypt4096(b *testing.B) {
	p := scrypt.Params{N: 4096, R: 8, P: 1, SaltLen: 8, DKLen: 32}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = scrypt.GenerateFromPassword([]byte("test"), p)
		}
	})
	b.ReportAllocs()
}
func BenchmarkScrypt16384(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = scrypt.GenerateFromPassword([]byte("test"), scrypt.DefaultParams)
		}
	})
	b.ReportAllocs()
}
func BenchmarkPbkdf2(b *testing.B) {
	salt := make([]byte, 8)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = pbkdf2.Key([]byte("test"), salt, 1024, 32, sha512.New)
		}
	})
	b.ReportAllocs()
}

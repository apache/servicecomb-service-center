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
package etcdsync_test

import (
	. "github.com/ServiceComb/service-center/pkg/etcdsync"

	"fmt"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Mutex", func() {
	Context("normal", func() {
		It("TestLockTimeout", func() {
			m1 := New("key1", 10)
			m2 := New("key1", 2)
			m1.Lock()
			fmt.Println("UT===================m1 locked")
			ch := make(chan bool)
			go func() {
				l, _ := m2.Lock()
				fmt.Println("UT===================m2 locked")
				l.Unlock()
				ch <- true
			}()
			<-ch
			fmt.Println("lock m1 timeout")
			m3 := New("key1", 2)
			l, _ := m3.Lock()
			fmt.Println("UT===================m3 locked")
			l.Unlock()

		})
	})
})

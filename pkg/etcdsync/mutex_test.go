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
package etcdsync

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mutex", func() {
	Context("normal", func() {
		It("TestLockTimeout", func() {
			m1 := NewLockFactory("key1", 5)
			m2 := NewLockFactory("key1", 1)
			l1, err := m1.NewDLock(true)
			Expect(l1).ToNot(BeNil())
			Expect(err).To(BeNil())

			fmt.Println("UT===================m1 locked")
			ch := make(chan bool)
			go func() {
				l2, err := m2.NewDLock(false)
				Expect(l2).To(BeNil())
				Expect(err).ToNot(BeNil())
				fmt.Println("UT===================m2 try failed")

				l2, err = m2.NewDLock(true) // 1s * 3
				Expect(l2).To(BeNil())
				Expect(err).ToNot(BeNil())
				fmt.Println("UT===================m2 timed out")
				ch <- true
			}()
			<-ch

			m3 := NewLockFactory("key1", 2)
			l3, err := m3.NewDLock(true)
			Expect(l3).ToNot(BeNil())
			Expect(err).To(BeNil())

			fmt.Println("UT===================m3 locked")
			err = l3.Unlock()
			Expect(err).To(BeNil())

			err = l1.Unlock()
			Expect(err).To(BeNil())
			fmt.Println("UT===================m1 unlocked")
		})
	})
})

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
package core_test

import (
	"..//lager"
	"..//lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("ReconfigurableSink", func() {
	var (
		testSink *lagertest.TestSink

		sink *lager.ReconfigurableSink
	)

	BeforeEach(func() {
		testSink = lagertest.NewTestSink()

		sink = lager.NewReconfigurableSink(testSink, lager.INFO)
	})

	It("returns the current level", func() {
		Ω(sink.GetMinLevel()).Should(Equal(lager.INFO))
	})

	Context("when logging above the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.INFO, []byte("hello world"))
		})

		It("writes to the given sink", func() {
			Ω(testSink.Buffer()).Should(gbytes.Say("hello world\n"))
		})
	})

	Context("when logging below the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.DEBUG, []byte("hello world"))
		})

		It("does not write to the given writer", func() {
			Ω(testSink.Buffer().Contents()).Should(BeEmpty())
		})
	})

	Context("when reconfigured to a new log level", func() {
		BeforeEach(func() {
			sink.SetMinLevel(lager.DEBUG)
		})

		It("writes logs above the new log level", func() {
			sink.Log(lager.DEBUG, []byte("hello world"))
			Ω(testSink.Buffer()).Should(gbytes.Say("hello world\n"))
		})

		It("returns the newly updated level", func() {
			Ω(sink.GetMinLevel()).Should(Equal(lager.DEBUG))
		})
	})
})

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
package cron_test

import (
	. "github.com/servicecomb/service-center/common/cron"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func DummyTask() {
	//
}

var _ = Describe("Cron", func() {
	Describe("cron every test", func() {
		It("Test every second cron", func() {
			job := Every(1).Seconds()
			job.Do(DummyTask)
			Expect(job.String()).To(Equal("[JOB]type: 1(seconds), interval: 1, delay: 0, at: , task: servicecenter/common/cron_test.DummyTask"))
		})

		It("Test every minute cron", func() {
			job := Every(3).Minutes()
			job.Do(DummyTask)
			Expect(job.String()).To(Equal("[JOB]type: 2(minutes), interval: 3, delay: 0, at: , task: servicecenter/common/cron_test.DummyTask"))
		})

		It("Test every hour cron", func() {
			job := Every(5).Hours()
			job.Do(DummyTask)
			Expect(job.String()).To(Equal("[JOB]type: 3(hours), interval: 5, delay: 0, at: , task: servicecenter/common/cron_test.DummyTask"))
		})

		It("Test every day cron", func() {
			job := Every(7).Days()
			job.Do(DummyTask)
			Expect(job.String()).To(Equal("[JOB]type: 4(days), interval: 7, delay: 0, at: , task: servicecenter/common/cron_test.DummyTask"))
		})
	})

	Describe("cron every with delay test", func() {
		It("Test every second with delay cron", func() {
			job := EveryWithDelay(1, 2).Seconds()
			job.Do(DummyTask)
			Expect(job.String()).To(Equal("[JOB]type: 1(seconds), interval: 1, delay: 2, at: , task: servicecenter/common/cron_test.DummyTask"))
		})

		It("Test every minutewith delay cron", func() {
			job := EveryWithDelay(3, 4).Minutes()
			job.Do(DummyTask)
			Expect(job.String()).To(Equal("[JOB]type: 2(minutes), interval: 3, delay: 4, at: , task: servicecenter/common/cron_test.DummyTask"))
		})

		It("Test every hourwith delay cron", func() {
			job := EveryWithDelay(5, 6).Hours()
			job.Do(DummyTask)
			Expect(job.String()).To(Equal("[JOB]type: 3(hours), interval: 5, delay: 6, at: , task: servicecenter/common/cron_test.DummyTask"))
		})

		It("Test every daywith delay cron", func() {
			job := EveryWithDelay(7, 8).Days()
			job.Do(DummyTask)
			Expect(job.String()).To(Equal("[JOB]type: 4(days), interval: 7, delay: 8, at: , task: servicecenter/common/cron_test.DummyTask"))
		})
	})

	Describe("cron at time test", func() {
		It("Test at time cron", func() {
			job := Every(1).Days().At("01:23")
			job.Do(DummyTask)
			Expect(job.String()).To(Equal("[JOB]type: 4(days), interval: 1, delay: 0, at: 01:23, task: servicecenter/common/cron_test.DummyTask"))
		})

		It("Test at time format invalid", func() {
			job := Every(1).Days().At("0123")
			Expect(job).To(BeNil())
		})

		It("Test at time invalid", func() {
			job := Every(1).Days().At("01:78")
			Expect(job).To(BeNil())
		})
	})

})

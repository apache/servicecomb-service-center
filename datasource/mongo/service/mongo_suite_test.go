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

package service_test

// initialize
import (
	"context"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/util"
	_ "github.com/apache/servicecomb-service-center/test"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"

	"testing"
	"time"

	. "github.com/onsi/gomega"
)

var timeLimit = 2 * time.Second

var _ = BeforeSuite(func() {
	//clear service created in last test
	time.Sleep(timeLimit)
	_ = datasource.Instance().ClearNoInstanceServices(context.Background(), timeLimit)
})

func getContext() context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), "default", "default"))
}

func TestMongo(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("mongo.junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "mongo Suite", []Reporter{junitReporter})
}

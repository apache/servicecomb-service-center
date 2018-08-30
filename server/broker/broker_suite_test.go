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
package broker_test

import _ "github.com/apache/incubator-servicecomb-service-center/server/init"
import _ "github.com/apache/incubator-servicecomb-service-center/server/bootstrap"
import (
	"github.com/apache/incubator-servicecomb-service-center/server/broker"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/service"
	"github.com/astaxie/beego"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"testing"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
}

var brokerResource = broker.BrokerServiceAPI

var _ = BeforeSuite(func() {
	//init plugin
	core.ServerInfo.Config.EnableCache = false
	core.ServiceAPI, core.InstanceAPI = service.AssembleResources()
})

func TestBroker(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("model.junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "model Suite", []Reporter{junitReporter})
}

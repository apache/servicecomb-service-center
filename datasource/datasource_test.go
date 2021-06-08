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

package datasource_test

import (
	_ "github.com/apache/servicecomb-service-center/test"

	"context"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

// Run The UT-TEST By Set The TEST_MODE ENV
// Mongo: TEST_MODE=mongo.
// ETCD: TEST_MODE=etcd.

func TestingContext() context.Context {
	return context.WithValue(context.Background(), util.CtxNocache, "1")
}

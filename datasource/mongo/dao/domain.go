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

package dao

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func AddDomain(ctx context.Context, domain string) error {
	d := Domain{
		Domain: domain,
	}
	result, err := client.GetMongoClient().Insert(ctx, CollectionDomain, d)
	if err == nil {
		log.Info(fmt.Sprintf("insert domain to mongodb success %s", result.InsertedID))
	}
	return err
}

func ExistDomain(ctx context.Context, filter interface{}) (bool, error) {
	return client.GetMongoClient().DocExist(ctx, CollectionDomain, filter)
}

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

db.createUser(
    {
        user: "sc",
        pwd: "123",
        roles: [
            {
                role: "readWrite",
                db: "servicecenter"
            }
        ]
    }
);

//index
db.service.createIndex({"service.service_id": 1}, {unique: true});
db.service.createIndex({
    "service.app": 1,
    "service.service_name": 1,
    "service.env": 1,
    "service.version": 1,
    "domain": 1,
    "project": 1,
}, {unique: true});
db.instance.createIndex({"refresh_time": 1}, {expireAfterSeconds: 60});
db.instance.createIndex({"instance.service_id": 1});
db.schema.createIndex({"domain": 1, "project": 1, "service_id": 1});
db.rule.createIndex({"domain": 1, "project": 1, "service_id": 1});
db.dependency.createIndex({"domain": 1, "project": 1, "service_key": 1});

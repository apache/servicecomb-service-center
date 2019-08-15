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
package config

import "testing"

func TestDefaultVerification(t *testing.T) {
	conf := DefaultConfig()
	conf.Verify()
}

func TestBindAddrVerification(t *testing.T) {
	conf := DefaultConfig()
	conf.BindAddr = "abcde"
	conf.Verify()
}

func TestRPCAddrAddrVerification(t *testing.T) {
	conf := DefaultConfig()
	conf.RPCAddr = "abcde"
	conf.Verify()
}

func TestLocalBindAddrVerification(t *testing.T) {
	conf := DefaultConfig()
	conf.BindAddr = "127.0.0.1"
	conf.Verify()
}

func TestLocalRPCAddrVerification(t *testing.T) {
	conf := DefaultConfig()
	conf.RPCAddr = "127.0.0.1"
	conf.Verify()
}

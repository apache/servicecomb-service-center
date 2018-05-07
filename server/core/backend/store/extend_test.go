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
package store

import "testing"

type extend struct {
}

func (e *extend) Name() string {
	return "test"
}

func (e *extend) Prefix() string {
	return "/test"
}

func (e *extend) InitSize() int {
	return 0
}

func TestInstallType(t *testing.T) {
	id, err := InstallType(&extend{})
	if err != nil {
		t.Fatal(err)
	}
	if id == NONEXIST {
		t.Fatal(err)
	}

	id, err = InstallType(&extend{})
	if id != NONEXIST || err == nil {
		t.Fatal("InstallType fail", err)
	}
}

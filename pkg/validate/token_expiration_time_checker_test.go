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

package validate

import (
	"testing"
)

func TestExpirationTime_MatchString(t *testing.T) {
	checker := TokenExpirationTimeChecker{}
	if !checker.MatchString("30m") {
		t.Fail()
	}
	if checker.MatchString("00s") {
		t.Fatalf("0s is not between 15m and 24h")
	}
	if checker.MatchString("14m59s") {
		t.Fatalf("0s is not between 15m and 24h")
	}
	if checker.MatchString("24h1s") {
		t.Fatalf("0s is not between 15m and 24h")
	}

}

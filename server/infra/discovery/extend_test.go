// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discovery

import "testing"

func TestNewAddOn(t *testing.T) {
	id, err := Install(NewAddOn("TestNewAddOn", nil))
	if id != TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = Install(NewAddOn("", Configure()))
	if id != TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = Install(nil)
	if id != TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = Install(NewAddOn("TestNewAddOn", Configure()))
	if id == TypeError || err != nil {
		t.Fatalf("TestNewAddOn failed")
	}
	_, err = Install(NewAddOn("TestNewAddOn", Configure()))
	if err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
}

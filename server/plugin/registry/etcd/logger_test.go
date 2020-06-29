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
package etcd

import "testing"

func TestClientLogger_Print(t *testing.T) {
	l := &clientLogger{}

	defer func() {
		recover()
		defer func() {
			recover()
			defer func() {
				recover()
			}()
			l.Fatalln("a", "b")
		}()
		l.Fatalf("%s", "b")
	}()
	l.Info("a", "b")
	l.Infof("%s", "b")
	l.Infoln("a", "b")
	l.Warning("a", "b")
	l.Warningf("%s", "b")
	l.Warningln("a", "b")
	l.Error("a", "b")
	l.Errorf("%s", "b")
	l.Errorln("a", "b")
	l.Print("a", "b")
	l.Printf("%s", "b")
	l.Println("a", "b")

	l.Format("a/b", 1, 0, "c")
	l.Format("a/b", 4, 0, "c")

	if !l.V(0) {
		t.Fatalf("TestClientLogger_Print")
	}

	l.Fatal("a", "b")
}

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
package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExist(t *testing.T) {
	if IsDirExist("./test/file") {
		t.Error("dir exist failed")
	}

	if IsFileExist("./test/file") {
		t.Error("file exist failed")
	}
}

func TestOpenFile(t *testing.T) {
	fileName := "file"
	f, err := OpenFile(fileName)
	if err != nil {
		t.Errorf("open file failed: %s", err)
	}
	f.Close()

	f, err = OpenFile(fileName)
	if err != nil {
		t.Errorf("open file failed: %s", err)
	}
	f.Close()
	os.RemoveAll(fileName)

	fileName = "./test/file"
	f, err = OpenFile(fileName)
	if err != nil {
		t.Logf("open file failed: %s", err)
	}
	f.Close()
	os.RemoveAll(filepath.Dir(fileName))
}

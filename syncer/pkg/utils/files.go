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
)

// IsDirExist checks if a dir exists
func IsDirExist(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir() || os.IsExist(err)
}

// IsFileExist checks if a file exists
func IsFileExist(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir() || os.IsExist(err)
}

// OpenFile if file not exist auto create
func OpenFile(path string) (*os.File, error)  {
	if IsFileExist(path) {
		return os.Open(path)
	}

	dir := filepath.Dir(path)
	if !IsDirExist(dir) {
		err := os.MkdirAll(dir, 0666)
		if err != nil {
			return nil , err
		}
	}
	return os.Create(path)
}

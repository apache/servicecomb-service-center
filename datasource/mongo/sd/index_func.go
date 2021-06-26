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

package sd

type indexFunc func(interface{}) string

type indexCols struct {
	indexFuncs []indexFunc
}

var DepIndexCols *indexCols
var InstIndexCols *indexCols
var ServiceIndexCols *indexCols
var RuleIndexCols *indexCols

func NewIndexCols() *indexCols {
	return &indexCols{indexFuncs: make([]indexFunc, 0)}
}

func (i *indexCols) AddIndexFunc(f indexFunc) {
	i.indexFuncs = append(i.indexFuncs, f)
}

func (i *indexCols) GetIndexs(data interface{}) (res []string) {
	for _, f := range i.indexFuncs {
		index := f(data)
		if len(index) != 0 {
			res = append(res, index)
		}
	}
	return
}

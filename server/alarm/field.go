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

package alarm

import "fmt"

type Field struct {
	Key   string
	Value interface{}
}

func FieldBool(key string, v bool) Field {
	return Field{Key: key, Value: v}
}

func FieldString(key string, v string) Field {
	return Field{Key: key, Value: v}
}

func FieldInt64(key string, v int64) Field {
	return Field{Key: key, Value: v}
}

func FieldInt(key string, v int) Field {
	return Field{Key: key, Value: v}
}

func FieldFloat64(key string, v float64) Field {
	return Field{Key: key, Value: v}
}

func AdditionalContext(format string, args ...interface{}) Field {
	return FieldString(FieldAdditionalContext, fmt.Sprintf(format, args...))
}

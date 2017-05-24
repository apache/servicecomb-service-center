//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package core

import (
	"bytes"
	"encoding/json"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

func FormatLogLevel(x LogLevel) string {
	var level string
	switch x {
	case DEBUG:
		level = "debug"
	case INFO:
		level = "info"
	case WARN:
		level = "warn"
	case ERROR:
		level = "error"
	case FATAL:
		level = "fatal"
	}
	return level
}

func (x LogLevel) MarshalJSON() ([]byte, error) {
	// var level string
	var level = FormatLogLevel(x)
	return json.Marshal(level)
}

/*
func (x LogLevel) MarshalJSON() ([]byte, error) {
	var level string
	switch x {
	case DEBUG:
		level = "debug"
	case INFO:
		level = "info"
	case WARN:
		level = "warn"
	case ERROR:
		level = "error"
	case FATAL:
		level = "fatal"
	}
	return json.Marshal(level)
}
*/

type Data map[string]interface{}

type LogFormat struct {
	Timestamp string   `json:"timestamp"`
	Source    string   `json:"source"`
	Message   string   `json:"message"`
	LogLevel  LogLevel `json:"log_level"`
	Data      Data     `json:"data"`
	ProcessID int      `json:"process_id"`
	File      string   `json:"file"`
	LineNo    int      `json:"lineno"`
	Method    string   `json:"method"`
}

func prettyPrint(in []byte) []byte {
	var out bytes.Buffer
	err := json.Indent(&out, in, "", "\t")
	if err != nil {
		return in
	}
	return out.Bytes()
}
func (log LogFormat) ToJSON() ([]byte, error) {
	j, err := json.Marshal(log)
	return prettyPrint(j), err
}

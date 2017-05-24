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
package chug_test

import (
	"fmt"
	"reflect"

	"../..//lager/chug"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

func MatchLogEntry(entry chug.LogEntry) types.GomegaMatcher {
	return &logEntryMatcher{entry}
}

type logEntryMatcher struct {
	entry chug.LogEntry
}

func (m *logEntryMatcher) Match(actual interface{}) (success bool, err error) {
	actualEntry, ok := actual.(chug.LogEntry)
	if !ok {
		return false, fmt.Errorf("MatchLogEntry must be passed a chug.LogEntry.  Got:\n%s", format.Object(actual, 1))
	}

	return m.entry.LogLevel == actualEntry.LogLevel &&
		m.entry.Source == actualEntry.Source &&
		m.entry.Message == actualEntry.Message &&
		m.entry.Session == actualEntry.Session &&
		reflect.DeepEqual(m.entry.Error, actualEntry.Error) &&
		m.entry.Trace == actualEntry.Trace &&
		reflect.DeepEqual(m.entry.Data, actualEntry.Data), nil
}

func (m *logEntryMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to equal", m.entry)
}

func (m *logEntryMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to equal", m.entry)
}

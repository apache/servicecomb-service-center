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

import "sync/atomic"

type ReconfigurableSink struct {
	sink Sink

	minLogLevel int32
}

func NewReconfigurableSink(sink Sink, initialMinLogLevel LogLevel) *ReconfigurableSink {
	return &ReconfigurableSink{
		sink: sink,

		minLogLevel: int32(initialMinLogLevel),
	}
}

func (sink *ReconfigurableSink) Log(level LogLevel, log []byte) {
	minLogLevel := LogLevel(atomic.LoadInt32(&sink.minLogLevel))

	if level < minLogLevel {
		return
	}

	sink.sink.Log(level, log)
}

func (sink *ReconfigurableSink) SetMinLevel(level LogLevel) {
	atomic.StoreInt32(&sink.minLogLevel, int32(level))
}

func (sink *ReconfigurableSink) GetMinLevel() LogLevel {
	return LogLevel(atomic.LoadInt32(&sink.minLogLevel))
}

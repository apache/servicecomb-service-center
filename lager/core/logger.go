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
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
	"bytes"
	"strconv"
)

const STACK_TRACE_BUFFER_SIZE = 1024 * 100

type Logger interface {
	RegisterSink(Sink)
	Session(task string, data ...Data) Logger
	SessionName() string
	Debug(action string, data ...Data)
	Info(action string, data ...Data)
	Warn(action string, err error, data ...Data)
	Error(action string, err error, data ...Data)
	Fatal(action string, err error, data ...Data)
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(err error, format string, args ...interface{})
	Errorf(err error, format string, args ...interface{})
	Fatalf(err error, format string, args ...interface{})
	WithData(Data) Logger
}

type logger struct {
	component     string
	task          string
	sinks         []Sink
	sessionID     string
	nextSession   uint64
	data          Data
	logFormatText bool
}

func NewLoggerExt(component string, isFormatText bool) Logger {
	return &logger{
		component:     component,
		task:          component,
		sinks:         []Sink{},
		data:          Data{},
		logFormatText: isFormatText,
	}
}

func NewLogger(component string) Logger {
	return NewLoggerExt(component, true)
}

func (l *logger) RegisterSink(sink Sink) {
	l.sinks = append(l.sinks, sink)
}

func (l *logger) SessionName() string {
	return l.task
}

func (l *logger) Session(task string, data ...Data) Logger {
	sid := atomic.AddUint64(&l.nextSession, 1)

	var sessionIDstr string

	if l.sessionID != "" {
		sessionIDstr = fmt.Sprintf("%s.%d", l.sessionID, sid)
	} else {
		sessionIDstr = fmt.Sprintf("%d", sid)
	}

	return &logger{
		component: l.component,
		task:      fmt.Sprintf("%s.%s", l.task, task),
		sinks:     l.sinks,
		sessionID: sessionIDstr,
		data:      l.baseData(data...),
	}
}

func (l *logger) WithData(data Data) Logger {
	return &logger{
		component: l.component,
		task:      l.task,
		sinks:     l.sinks,
		sessionID: l.sessionID,
		data:      l.baseData(data),
	}
}

func (l *logger) log(loglevel LogLevel, action string, err error, data ...Data) {
	logData := l.baseData(data...)

	if err != nil {
		logData["error"] = err.Error()
	}

	if loglevel == FATAL {
		stackTrace := make([]byte, STACK_TRACE_BUFFER_SIZE)
		stackSize := runtime.Stack(stackTrace, false)
		stackTrace = stackTrace[:stackSize]

		logData["trace"] = string(stackTrace)
	}

	log := LogFormat{
		Timestamp: currentTimestamp(),
		Source:    l.component,
		Message:   fmt.Sprintf("%s.%s", l.task, strconv.QuoteToGraphic(action)),
		LogLevel:  loglevel,
		Data:      logData,
	}

	// add process_id, file, lineno, method to log data
	addExtLogInfo(&log)

	for _, sink := range l.sinks {
		if !(l.logFormatText) {
			jsondata, jserr := log.ToJSON()
			if jserr != nil {
				fmt.Printf("[lager] ToJSON() ERROR! action: %s, jserr: %s, log: %s\n", action, jserr, log)

				// also output json marshal error event to sink
				log.Data = Data{"Data": fmt.Sprint(logData)}
				jsonerrdata, _ := log.ToJSON()
				sink.Log(ERROR, jsonerrdata)

				continue
			}
			sink.Log(loglevel, jsondata)
		} else {
			levelstr := strings.ToUpper(FormatLogLevel(log.LogLevel))
			buf := bytes.Buffer{}
			buf.WriteString(fmt.Sprintf("%s %s %s %d %s %s():%d - %s",
				log.Timestamp, levelstr, log.Source, log.ProcessID, log.File, log.Method, log.LineNo, log.Message))
			if err != nil {
				buf.WriteString(fmt.Sprintf("(error: %s)", logData["error"]))
			}
			if loglevel == FATAL {
				buf.WriteString(fmt.Sprintf("(trace: %s)", logData["trace"]))
			}
			sink.Log(loglevel, buf.Bytes())
		}
	}

	if loglevel == FATAL {
		panic(err)
	}
}

func (l *logger) Debug(action string, data ...Data) {
	l.log(DEBUG, action, nil, data...)
}

func (l *logger) Info(action string, data ...Data) {
	l.log(INFO, action, nil, data...)
}

func (l *logger) Warn(action string, err error, data ...Data) {
	l.log(WARN, action, err, data...)
}

func (l *logger) Error(action string, err error, data ...Data) {
	l.log(ERROR, action, err, data...)
}

func (l *logger) Fatal(action string, err error, data ...Data) {
	l.log(FATAL, action, err, data...)
}

func (l *logger) logf(loglevel LogLevel, err error, format string, args ...interface{}) {
	logmsg := fmt.Sprintf(format, args...)
	l.log(loglevel, logmsg, err)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.logf(DEBUG, nil, format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.logf(INFO, nil, format, args...)
}

func (l *logger) Warnf(err error, format string, args ...interface{}) {
	l.logf(WARN, err, format, args...)
}

func (l *logger) Errorf(err error, format string, args ...interface{}) {
	l.logf(ERROR, err, format, args...)
}

func (l *logger) Fatalf(err error, format string, args ...interface{}) {
	l.logf(FATAL, err, format, args...)
}

func (l *logger) baseData(givenData ...Data) Data {
	data := Data{}

	for k, v := range l.data {
		data[k] = v
	}

	if len(givenData) > 0 {
		for _, dataArg := range givenData {
			for key, val := range dataArg {
				data[key] = val
			}
		}
	}

	if l.sessionID != "" {
		data["session"] = l.sessionID
	}

	return data
}

func currentTimestamp() string {
	return time.Now().Format(time.RFC3339Nano)
}

func addExtLogInfo(logf *LogFormat) {
	logf.ProcessID = os.Getpid()

	for i := 3; i <= 5; i++ {
		pc, file, line, ok := runtime.Caller(i)

		if strings.Index(file, "logger.go") > 0 {
			continue
		}

		if ok {
			idx := strings.LastIndex(file, "src")
			switch {
			case idx >= 0:
				logf.File = file[idx+4:]
			default:
				logf.File = file
			}

			logf.LineNo = line
			if f := runtime.FuncForPC(pc); f != nil {
				logf.Method = formatFuncName(f.Name())
			}
		}
		break
	}
}

func formatFuncName(f string) string {
	i := strings.LastIndex(f, "/")
	j := strings.Index(f[i+1:], ".")
	if j < 1 {
		return "???"
	}
	_, fun := f[:i+j+1], f[i+j+2:]
	i = strings.LastIndex(fun, ".")
	return fun[i+1:]
}

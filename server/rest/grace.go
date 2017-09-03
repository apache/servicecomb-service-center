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
package rest

import (
	"flag"
	"fmt"
	"github.com/ServiceComb/service-center/util"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const (
	PreSignal = iota
	PostSignal
)

var (
	isFork         bool
	filesOrder     string
	files          []*os.File
	filesOffsetMap map[string]int

	SignalHooks map[int]map[os.Signal][]func()
	graceMux    sync.Mutex
	forked      bool
)

func init() {
	flag.BoolVar(&isFork, "fork", false, "listen on open fd (after forking)")
	flag.StringVar(&filesOrder, "filesorder", "", "previous initialization FDs order")

	filesOffsetMap = make(map[string]int)
	SignalHooks = map[int]map[os.Signal][]func(){
		PreSignal: {
			syscall.SIGHUP:  {},
			syscall.SIGQUIT: {},
			syscall.SIGINT:  {},
			syscall.SIGTERM: {},
		},
		PostSignal: {
			syscall.SIGHUP:  {},
			syscall.SIGQUIT: {},
			syscall.SIGINT:  {},
			syscall.SIGTERM: {},
		},
	}
}

func InitGrace() {
	if !flag.Parsed() {
		flag.Parse()
	}

	go handleSignals()
}

func RegisterSignalHook(phase int, f func(), sigs ...os.Signal) {
	for s := range SignalHooks[phase] {
		for _, sig := range sigs {
			if s == sig {
				SignalHooks[phase][sig] = append(SignalHooks[phase][sig], f)
			}
		}
	}
}

func RegisterFiles(name string, f *os.File) {
	if f == nil {
		return
	}
	graceMux.Lock()
	filesOffsetMap[name] = len(files)
	files = append(files, f)
	graceMux.Unlock()
}

func fireSignalHook(ppFlag int, sig os.Signal) {
	if _, notSet := SignalHooks[ppFlag][sig]; !notSet {
		return
	}
	for _, f := range SignalHooks[ppFlag][sig] {
		f()
	}
}

func handleSignals() {
	var sig os.Signal

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig = <-sigCh
		fireSignalHook(PreSignal, sig)
		switch sig {
		case syscall.SIGHUP:
			util.Logger().Debugf("received signal 'SIGHUP', now forking")
			err := fork()
			if err != nil {
				util.Logger().Errorf(err, "fork a process failed")
			}
		default:
			util.Logger().Warnf(nil, "received signal '%v'", sig)
		}
		fireSignalHook(PostSignal, sig)
	}
}

func fork() (err error) {
	graceMux.Lock()
	defer graceMux.Unlock()
	if forked {
		return
	}
	forked = true

	var orderArgs = make([]string, len(filesOffsetMap))
	for name, i := range filesOffsetMap {
		orderArgs[i] = name
	}
	path := os.Args[0]
	var args []string
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if arg == "-fork" {
				break
			}
			args = append(args, arg)
		}
	}
	args = append(args, "-fork")
	if len(filesOffsetMap) > 0 {
		args = append(args, fmt.Sprintf(`-filesorder=%s`, strings.Join(orderArgs, ",")))
	}
	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = files
	err = cmd.Start()
	if err != nil {
		util.Logger().Errorf(err, "fork a process failed, %s %v", path, args)
		return
	}
	util.Logger().Warnf(nil, "fork process %s %v", path, args)
	return
}

func IsFork() bool {
	return isFork
}

func ExtraFileOrder(name string) int {
	if len(filesOrder) == 0 {
		return -1
	}
	orders := strings.Split(filesOrder, ",")
	for i, f := range orders {
		if f == name {
			return i
		}
	}
	return -1
}

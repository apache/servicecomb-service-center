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
package grace

import (
	"flag"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
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

	registerSignals []os.Signal
	SignalHooks     map[int]map[os.Signal][]func()
	graceMux        sync.Mutex
	forked          bool
)

func init() {
	flag.BoolVar(&isFork, "fork", false, "listen on open fd (after forking)")
	flag.StringVar(&filesOrder, "filesorder", "", "previous initialization FDs order")

	registerSignals = []os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM,
	}
	filesOffsetMap = make(map[string]int)
	SignalHooks = map[int]map[os.Signal][]func(){
		PreSignal:  {},
		PostSignal: {},
	}
	for _, sig := range registerSignals {
		SignalHooks[PreSignal][sig] = []func(){}
		SignalHooks[PostSignal][sig] = []func(){}
	}

	go handleSignals()
}

func ParseCommandLine() {
	if !flag.Parsed() {
		flag.Parse()
	}
}

func Before(f func()) {
	RegisterSignalHook(PreSignal, f, syscall.SIGHUP)
}

func After(f func()) {
	RegisterSignalHook(PostSignal, f, registerSignals[1:]...)
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
	signal.Notify(sigCh, registerSignals...)

	for {
		select {
		case sig = <-sigCh:
			fireSignalHook(PreSignal, sig)
			switch sig {
			case syscall.SIGHUP:
				log.Debugf("received signal '%v', now forking", sig)
				err := fork()
				if err != nil {
					log.Errorf(err, "fork a process failed")
				}
			}
			fireSignalHook(PostSignal, sig)
		}
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

	// add fork and file descriptions order flags
	args := append(parseCommandLine(), "-fork")
	if len(filesOffsetMap) > 0 {
		args = append(args, fmt.Sprintf(`-filesorder=%s`, strings.Join(orderArgs, ",")))
	}

	if err = newCommand(args...); err != nil {
		log.Errorf(err, "fork a process failed, %v", args)
		return
	}
	log.Warnf("fork process %v", args)
	return
}

func parseCommandLine() (args []string) {
	if len(os.Args) <= 1 {
		return
	}
	// ignore process path
	for _, arg := range os.Args[1:] {
		if arg == "-fork" {
			// ignore fork flags
			break
		}
		args = append(args, arg)
	}
	return
}

func newCommand(args ...string) error {
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = files
	return cmd.Start()
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

func Done() error {
	if !IsFork() {
		return nil
	}

	ppid := os.Getppid()
	process, err := os.FindProcess(ppid)
	if err != nil {
		return err
	}
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}
	return nil
}

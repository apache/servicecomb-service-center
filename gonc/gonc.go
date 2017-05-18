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
package main

import (
	"net"
	"fmt"
	"time"
	"flag"
	"os"
	"strings"
	"strconv"
)

const DEFAULT_TCP_CONNCET_TIMEOUT = 3

func tcpConnect(ip string, port int16, timeout int, verbose bool) error {
	address := fmt.Sprintf("%s:%d", ip, port)
	if verbose {
		fmt.Printf("Test connection to %s(timeout: %d)...\n", address, timeout)
	}
	conn, err := net.DialTimeout("tcp", address, time.Duration(timeout) * time.Second)
	if err != nil {
		fmt.Printf("connect to %s failed, err: %+v\n", address, err)
		return err
	}

	if verbose {
		fmt.Printf("Test connection to %s(timeout: %d) OK.\n", address, timeout)
	}

	conn.Close()
	return nil
}

// nc(go version)
// useage: gonc [-v] [-t timeout] ip port
func main() {
	flag.Bool("h", false, "print help message")
	pVerbose := flag.Bool("v", false, "Make the operation more talkative")
	pTimeout := flag.Int("w", DEFAULT_TCP_CONNCET_TIMEOUT, "connect timeout")

	flag.Parse()

	argv := len(os.Args)
	if argv == 1 || strings.ToLower(os.Args[1]) == "-h" || strings.ToLower(os.Args[1]) == "--help" {
		fmt.Println("gonc [-v] [-t timeout] ip port")
		flag.Usage()
		os.Exit(0)
	}

	timeout := *pTimeout
	verbose := *pVerbose
	ip := flag.Arg(0)
	port, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		fmt.Println("Invalid parameters.", err)
		os.Exit(-1)
	}

	err = tcpConnect(ip, int16(port), timeout, verbose)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

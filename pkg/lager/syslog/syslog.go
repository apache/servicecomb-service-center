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
package syslog

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

type Writer struct {
	conn net.Conn
}

var syslogHeader string

func New() (*Writer, error) {
	return Dial("", "", "", "")
}

func Dial(component, appguid, network, raddr string) (*Writer, error) {

	hostname, _ := os.Hostname()
	// construct syslog header the same to rsyslog's,
	// origin, node_id, app_guid, instance_id, loglevel
	syslogHeader = fmt.Sprintf("%s  %s  %s  %s  %s", component, component, appguid, hostname, "all")

	var conn net.Conn
	var err error
	if network == "" {
		conn, err = unixSyslog()
	} else {
		conn, err = net.Dial(network, raddr)
	}
	return &Writer{
		conn: conn,
	}, err
}

func (r *Writer) Write(b []byte) (int, error) {
	nl := ""
	if len(b) == 0 || b[len(b)-1] != '\n' {
		nl = "\n"
	}

	r.conn.SetWriteDeadline(time.Now().Add(1 * time.Second))

	_, err := fmt.Fprintf(r.conn, "  %s  %s%s", syslogHeader, b, nl)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (r *Writer) Close() error {
	return r.conn.Close()
}

func unixSyslog() (net.Conn, error) {
	logTypes := []string{"unixgram", "unix"}
	logPaths := []string{"/dev/log", "/var/run/syslog"}
	var raddr string
	for _, network := range logTypes {
		for _, path := range logPaths {
			raddr = path
			conn, err := net.Dial(network, raddr)
			if err != nil {
				continue
			} else {
				return conn, nil
			}
		}
	}
	return nil, errors.New("Could not connect to local syslog socket")
}

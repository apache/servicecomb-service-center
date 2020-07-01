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

package util

import (
	"github.com/prometheus/procfs"
	"os"
	"strconv"
	"unsafe"
)

const intSize = int(unsafe.Sizeof(0))

var bs *[intSize]byte

func init() {
	i := 0x1
	bs = (*[intSize]byte)(unsafe.Pointer(&i))
}

func IsBigEndian() bool {
	return !IsLittleEndian()
}

func IsLittleEndian() bool {
	return bs[0] == 0
}

func PathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func HostName() (hostname string) {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}
	return
}

func GetEnvInt(name string, def int) int {
	env, ok := os.LookupEnv(name)
	if ok {
		i64, err := strconv.ParseInt(env, 10, 0)
		if err != nil {
			return def
		}
		return int(i64)
	}
	return def
}

func GetEnvString(name string, def string) string {
	env, ok := os.LookupEnv(name)
	if ok {
		return env
	}
	return def
}

func GetProcCPUUsage() (pt float64, ct float64) {
	p, _ := procfs.NewProc(os.Getpid())
	stat, _ := procfs.NewStat()
	pstat, _ := p.NewStat()
	ct = stat.CPUTotal.User + stat.CPUTotal.Nice + stat.CPUTotal.System +
		stat.CPUTotal.Idle + stat.CPUTotal.Iowait + stat.CPUTotal.IRQ +
		stat.CPUTotal.SoftIRQ + stat.CPUTotal.Steal + stat.CPUTotal.Guest
	pt = float64(pstat.UTime+pstat.STime+pstat.CUTime+pstat.CSTime) / 100
	return
}

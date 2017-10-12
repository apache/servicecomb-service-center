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
package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"net"
	"sync"
	"time"
	"unsafe"
)

// UUID layout variants.
const (
	LAYOUT_NCS = iota
	LAYOUT_RFC4122
	LAYOUT_MICROSOFT
	LAYOUT_FUTURE
)

// Difference in 100-nanosecond intervals between
// UUID epoch (October 15, 1582) and Unix epoch (January 1, 1970).
const EPOCH_START_IN_NS = 122192928000000000

const DASH byte = '-'

var (
	mux          sync.Mutex
	once         sync.Once
	epochFunc    = unixTimeFunc
	random       uint16
	last         uint64
	hardwareAddr [6]byte
)

func initRandom() {
	buf := make([]byte, 2)
	doSafeRandom(buf)
	random = binary.BigEndian.Uint16(buf)
}

func initHardwareAddr() {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if len(iface.HardwareAddr) >= 6 {
				copy(hardwareAddr[:], iface.HardwareAddr)
				return
			}
		}
	}

	// Initialize hardwareAddr randomly in case
	// of real network interfaces absence
	doSafeRandom(hardwareAddr[:])

	// Set multicast bit as recommended in RFC 4122
	hardwareAddr[0] |= 0x01
}

func initialize() {
	initRandom()
	initHardwareAddr()
}

func doSafeRandom(dest []byte) {
	if _, err := rand.Read(dest); err != nil {
		panic(err)
	}
}

func unixTimeFunc() uint64 {
	return EPOCH_START_IN_NS + uint64(time.Now().UnixNano()/100)
}

type UUID [16]byte

func (u UUID) Bytes() []byte {
	return u[:]
}

// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (u UUID) String() string {
	buf := make([]byte, 36)

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = DASH
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = DASH
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = DASH
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = DASH
	hex.Encode(buf[24:], u[10:])

	return *(*string)(unsafe.Pointer(&buf))
}

func (u *UUID) SetVersion(v byte) {
	u[6] = (u[6] & 0x0f) | (v << 4)
}

func (u UUID) Version() uint {
	return uint(u[6] >> 4)
}

func (u *UUID) SetLayout() {
	u[8] = (u[8] & 0xbf) | 0x80 // LAYOUT_RFC4122
}

func (u UUID) Layout() uint {
	switch {
	case (u[8] & 0x80) == 0x00:
		return LAYOUT_NCS
	case (u[8]&0xc0)|0x80 == 0x80:
		return LAYOUT_RFC4122
	case (u[8]&0xe0)|0xc0 == 0xc0:
		return LAYOUT_MICROSOFT
	}
	return LAYOUT_FUTURE
}

func doInit() (uint64, uint16, []byte) {
	once.Do(initialize)

	mux.Lock()

	now := epochFunc()
	if now <= last {
		random++
	}
	last = now

	mux.Unlock()
	return now, random, hardwareAddr[:]
}

func NewV1() UUID {
	u := UUID{}

	timeNow, clockSeq, hardwareAddr := doInit()

	binary.BigEndian.PutUint32(u[0:], uint32(timeNow))
	binary.BigEndian.PutUint16(u[4:], uint16(timeNow>>32))
	binary.BigEndian.PutUint16(u[6:], uint16(timeNow>>48))
	binary.BigEndian.PutUint16(u[8:], clockSeq)

	copy(u[10:], hardwareAddr)

	u.SetVersion(1)
	u.SetLayout()

	return u
}

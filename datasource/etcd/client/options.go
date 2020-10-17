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

package client

import (
	"bytes"
	"fmt"
)

//Options contains configuration for plugins
type Options struct {
	PluginImplName ImplName
}

type PluginOp struct {
	Action        ActionType
	Key           []byte
	EndKey        []byte
	Value         []byte
	Prefix        bool
	PrevKV        bool
	Lease         int64
	KeyOnly       bool
	CountOnly     bool
	SortOrder     SortOrder
	Revision      int64
	IgnoreLease   bool
	Mode          CacheMode
	WatchCallback WatchCallback
	Offset        int64
	Limit         int64
	Global        bool
}

func (op PluginOp) String() string {
	return op.FormatURLParams()
}

func (op PluginOp) FormatURLParams() string {
	var buf bytes.Buffer
	buf.WriteString("action=")
	buf.WriteString(op.Action.String())
	buf.WriteString("&mode=")
	buf.WriteString(op.Mode.String())
	buf.WriteString("&key=")
	buf.Write(op.Key)
	buf.WriteString(fmt.Sprintf("&len=%d", len(op.Value)))
	if len(op.EndKey) > 0 {
		buf.WriteString("&end=")
		buf.Write(op.EndKey)
	}
	if op.Prefix {
		buf.WriteString("&prefix=true")
	}
	if op.PrevKV {
		buf.WriteString("&prev=true")
	}
	if op.Lease > 0 {
		buf.WriteString(fmt.Sprintf("&lease=%d", op.Lease))
	}
	if op.KeyOnly {
		buf.WriteString("&keyOnly=true")
	}
	if op.CountOnly {
		buf.WriteString("&countOnly=true")
	}
	if op.SortOrder != SortNone {
		buf.WriteString("&sort=")
		buf.WriteString(op.SortOrder.String())
	}
	if op.Revision > 0 {
		buf.WriteString(fmt.Sprintf("&rev=%d", op.Revision))
	}
	if op.IgnoreLease {
		buf.WriteString("&ignoreLease=true")
	}
	if op.Offset > 0 {
		buf.WriteString(fmt.Sprintf("&offset=%d", op.Offset))
	}
	if op.Limit > 0 {
		buf.WriteString(fmt.Sprintf("&limit=%d", op.Limit))
	}
	if op.Global {
		buf.WriteString("&global=true")
	}
	return buf.String()
}

func (op PluginOp) NoCache() bool {
	return op.Mode == ModeNoCache ||
		op.Revision > 0 ||
		(op.Offset >= 0 && op.Limit > 0)
}

func (op PluginOp) CacheOnly() bool {
	return op.Mode == ModeCache
}

type Operation func(...PluginOpOption) (op PluginOp)

type PluginOpOption func(*PluginOp)
type WatchCallback func(message string, evt *PluginResponse) error

var GET PluginOpOption = func(op *PluginOp) { op.Action = Get }
var PUT PluginOpOption = func(op *PluginOp) { op.Action = Put }
var DEL PluginOpOption = func(op *PluginOp) { op.Action = Delete }

func WithKey(key []byte) PluginOpOption      { return func(op *PluginOp) { op.Key = key } }
func WithEndKey(key []byte) PluginOpOption   { return func(op *PluginOp) { op.EndKey = key } }
func WithValue(value []byte) PluginOpOption  { return func(op *PluginOp) { op.Value = value } }
func WithPrefix() PluginOpOption             { return func(op *PluginOp) { op.Prefix = true } }
func WithPrevKv() PluginOpOption             { return func(op *PluginOp) { op.PrevKV = true } }
func WithLease(leaseID int64) PluginOpOption { return func(op *PluginOp) { op.Lease = leaseID } }
func WithKeyOnly() PluginOpOption            { return func(op *PluginOp) { op.KeyOnly = true } }
func WithCountOnly() PluginOpOption          { return func(op *PluginOp) { op.CountOnly = true } }
func WithGlobal() PluginOpOption             { return func(op *PluginOp) { op.Global = true } }
func WithNoneOrder() PluginOpOption          { return func(op *PluginOp) { op.SortOrder = SortNone } }
func WithAscendOrder() PluginOpOption        { return func(op *PluginOp) { op.SortOrder = SortAscend } }
func WithDescendOrder() PluginOpOption       { return func(op *PluginOp) { op.SortOrder = SortDescend } }
func WithRev(revision int64) PluginOpOption  { return func(op *PluginOp) { op.Revision = revision } }
func WithIgnoreLease() PluginOpOption        { return func(op *PluginOp) { op.IgnoreLease = true } }
func WithCacheOnly() PluginOpOption          { return func(op *PluginOp) { op.Mode = ModeCache } }
func WithNoCache() PluginOpOption            { return func(op *PluginOp) { op.Mode = ModeNoCache } }
func WithWatchCallback(f WatchCallback) PluginOpOption {
	return func(op *PluginOp) { op.WatchCallback = f }
}
func WithStrKey(key string) PluginOpOption     { return WithKey([]byte(key)) }
func WithStrEndKey(key string) PluginOpOption  { return WithEndKey([]byte(key)) }
func WithStrValue(value string) PluginOpOption { return WithValue([]byte(value)) }
func WithOffset(i int64) PluginOpOption        { return func(op *PluginOp) { op.Offset = i } }
func WithLimit(i int64) PluginOpOption         { return func(op *PluginOp) { op.Limit = i } }
func WatchPrefixOpOptions(key string) []PluginOpOption {
	return []PluginOpOption{GET, WithStrKey(key), WithPrefix(), WithPrevKv()}
}

func OpGet(opts ...PluginOpOption) (op PluginOp) {
	op = OptionsToOp(opts...)
	op.Action = Get
	return
}
func OpPut(opts ...PluginOpOption) (op PluginOp) {
	op = OptionsToOp(opts...)
	op.Action = Put
	return
}
func OpDel(opts ...PluginOpOption) (op PluginOp) {
	op = OptionsToOp(opts...)
	op.Action = Delete
	return
}
func OptionsToOp(opts ...PluginOpOption) (op PluginOp) {
	for _, opt := range opts {
		opt(&op)
	}
	if op.Limit == 0 {
		op.Offset = -1
		op.Limit = DefaultPageCount
	}
	return
}

type CompareOp struct {
	Key    []byte
	Type   CompareType
	Result CompareResult
	Value  interface{}
}

func (op CompareOp) String() string {
	return fmt.Sprintf(
		"{key: %s, type: %s, result: %s, val: %s}",
		op.Key, op.Type, op.Result, op.Value,
	)
}

type CompareOperation func(op *CompareOp)

func CmpVer(key []byte) CompareOperation {
	return func(op *CompareOp) { op.Key = key; op.Type = CmpVersion }
}
func CmpCreateRev(key []byte) CompareOperation {
	return func(op *CompareOp) { op.Key = key; op.Type = CmpCreate }
}
func CmpModRev(key []byte) CompareOperation {
	return func(op *CompareOp) { op.Key = key; op.Type = CmpMod }
}
func CmpVal(key []byte) CompareOperation {
	return func(op *CompareOp) { op.Key = key; op.Type = CmpValue }
}
func CmpStrVer(key string) CompareOperation       { return CmpVer([]byte(key)) }
func CmpStrCreateRev(key string) CompareOperation { return CmpCreateRev([]byte(key)) }
func CmpStrModRev(key string) CompareOperation    { return CmpModRev([]byte(key)) }
func CmpStrVal(key string) CompareOperation       { return CmpVal([]byte(key)) }
func OpCmp(opt CompareOperation, result CompareResult, v interface{}) (cmp CompareOp) {
	opt(&cmp)
	cmp.Result = result
	cmp.Value = v
	return cmp
}

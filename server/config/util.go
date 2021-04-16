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

package config

import (
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/go-chassis/go-archaius"
	"github.com/iancoleman/strcase"
)

func newOptions(key string, opts []Option) *Options {
	options := NewOptions(opts...)
	key = strings.ReplaceAll(key, ".", "_")
	if options.ENV == "" {
		options.ENV = strcase.ToScreamingSnake(key)
	}
	if options.Standby == "" {
		options.Standby = strcase.ToSnake(key)
	}
	return options
}

// GetString return the string type value by specified key
func GetString(key, def string, opts ...Option) string {
	options := newOptions(key, opts)
	if archaius.Exist(options.ENV) {
		return strings.TrimSpace(archaius.GetString(options.ENV, def))
	}
	if archaius.Exist(key) {
		return strings.TrimSpace(archaius.GetString(key, def))
	}
	return strings.TrimSpace(beego.AppConfig.DefaultString(options.Standby, def))
}

// GetBool return the boolean type value by specified key
func GetBool(key string, def bool, opts ...Option) bool {
	options := newOptions(key, opts)
	if archaius.Exist(options.ENV) {
		return archaius.GetBool(options.ENV, def)
	}
	if archaius.Exist(key) {
		return archaius.GetBool(key, def)
	}
	return beego.AppConfig.DefaultBool(options.Standby, def)
}

// GetInt return the int type value by specified key
func GetInt(key string, def int, opts ...Option) int {
	options := newOptions(key, opts)
	if archaius.Exist(options.ENV) {
		return archaius.GetInt(options.ENV, def)
	}
	if archaius.Exist(key) {
		return archaius.GetInt(key, def)
	}
	if archaius.Exist(options.Standby) {
		return archaius.GetInt(options.Standby, def)
	}
	return beego.AppConfig.DefaultInt(options.Standby, def)
}

// GetInt64 return the int64 type value by specified key
func GetInt64(key string, def int64, opts ...Option) int64 {
	options := newOptions(key, opts)
	if archaius.Exist(options.ENV) {
		return archaius.GetInt64(options.ENV, def)
	}
	if archaius.Exist(key) {
		return archaius.GetInt64(key, def)
	}
	return beego.AppConfig.DefaultInt64(options.Standby, def)
}

// GetDuration return the time.Duration type value by specified key
func GetDuration(key string, def time.Duration, opts ...Option) time.Duration {
	str := strings.TrimSpace(GetString(key, "", opts...))
	if str == "" {
		return def
	}
	d, err := time.ParseDuration(str)
	if err != nil {
		return def
	}
	return d
}

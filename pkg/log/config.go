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

package log

// Config struct for lager and rotate parameters
type Config struct {
	LoggerLevel string
	LoggerFile  string
	// if false, log print with JSON format
	LogFormatText bool
	// M Bytes
	LogRotateSize  int
	LogBackupCount int
	// days
	LogBackupAge int
	CallerSkip   int
	NoTime       bool // if true, not record time
	NoLevel      bool // if true, not record level
	NoCaller     bool // if true, not record caller
	// Event driven
	FlushFunc   func()
	RecoverFunc func(r interface{})
}

func (cfg Config) WithCallerSkip(s int) Config {
	cfg.CallerSkip = s
	return cfg
}

func (cfg Config) WithFile(path string) Config {
	cfg.LoggerFile = path
	return cfg
}

func (cfg Config) WithNoTime(b bool) Config {
	cfg.NoTime = b
	return cfg
}

func (cfg Config) WithNoLevel(b bool) Config {
	cfg.NoLevel = b
	return cfg
}

func (cfg Config) WithNoCaller(b bool) Config {
	cfg.NoCaller = b
	return cfg
}

func (cfg Config) WithExitFunc(f func()) Config {
	cfg.FlushFunc = f
	return cfg
}

func (cfg Config) WithRecoverFunc(f func(itf interface{})) Config {
	cfg.RecoverFunc = f
	return cfg
}

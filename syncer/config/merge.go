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
	"time"
)

// Merge multiple configurations into one
func Merge(configs ...Config) (conf Config) {
	for _, config := range configs {
		conf = merge(conf, config)
	}
	return
}

func merge(src, dst Config) Config {
	src.Mode = mergeString(src.Mode, dst.Mode)
	src.Node = mergeString(src.Node, dst.Node)
	src.Cluster = mergeString(src.Cluster, dst.Cluster)
	src.DataDir = mergeString(src.DataDir, dst.DataDir)

	src.Listener.BindAddr = mergeString(src.Listener.BindAddr, dst.Listener.BindAddr)
	src.Listener.RPCAddr = mergeString(src.Listener.RPCAddr, dst.Listener.RPCAddr)
	src.Listener.PeerAddr = mergeString(src.Listener.PeerAddr, dst.Listener.PeerAddr)
	src.Listener.TLSMount.Enabled = mergeBool(src.Listener.TLSMount.Enabled, dst.Listener.TLSMount.Enabled)
	src.Listener.TLSMount.Name = mergeString(src.Listener.TLSMount.Name, dst.Listener.TLSMount.Name)

	src.Join.Enabled = mergeBool(src.Join.Enabled, dst.Join.Enabled)
	src.Join.Address = mergeString(src.Join.Address, dst.Join.Address)
	src.Join.RetryMax = mergeInt(src.Join.RetryMax, dst.Join.RetryMax)
	src.Join.RetryInterval = mergeTimeString(src.Join.RetryInterval, dst.Join.RetryInterval)

	src.Task.Kind = mergeString(src.Task.Kind, dst.Task.Kind)
	src.Task.Params = mergeLabels(src.Task.Params, dst.Task.Params)

	src.Registry.Address = mergeString(src.Registry.Address, dst.Registry.Address)
	src.Registry.Plugin = mergeString(src.Registry.Plugin, dst.Registry.Plugin)
	src.Registry.TLSMount.Enabled = mergeBool(src.Registry.TLSMount.Enabled, dst.Registry.TLSMount.Enabled)
	src.Registry.TLSMount.Name = mergeString(src.Registry.TLSMount.Name, dst.Registry.TLSMount.Name)

	src.HttpConfig.HttpAddr = mergeString(src.HttpConfig.HttpAddr, dst.HttpConfig.HttpAddr)
	src.HttpConfig.Compressed = mergeBool(src.HttpConfig.Compressed, dst.HttpConfig.Compressed)
	src.HttpConfig.CompressMinBytes = mergeInt(src.HttpConfig.CompressMinBytes, dst.HttpConfig.CompressMinBytes)

	src.TLSConfigs = mergeTLSConfigs(src.TLSConfigs, dst.TLSConfigs)
	return src
}

func mergeString(src, dst string) string {
	if dst != "" {
		return dst
	}
	return src
}

func mergeInt(src, dst int) int {
	if dst != 0 {
		return dst
	}
	return src
}

func mergeBool(src, dst bool) bool {
	return dst
}

func mergeTimeString(src, dst string) string {
	_, err := time.ParseDuration(dst)
	if err != nil {
		return src
	}
	return dst
}

func mergeLabels(src, dst []Label) []Label {
	if len(src) == 0 {
		return dst[:]
	}

	merges := src[:]
	for _, dv := range dst {
		if findInLabels(src, dv.Key) == nil {
			merges = append(merges, dv)
		}
	}
	return merges
}

func findInLabels(labels []Label, key string) *Label {
	for _, item := range labels {
		if item.Key == key {
			return &item
		}
	}
	return nil
}

func mergeTLSConfigs(src, dst []*TLSConfig) []*TLSConfig {
	if len(src) == 0 {
		return dst[:]
	}

	merges := src[:]
	for _, dv := range dst {
		if findInTLSConfigs(src, dv.Name) == nil {
			merges = append(merges, dv)
		}
	}
	return merges
}

func findInTLSConfigs(list []*TLSConfig, name string) *TLSConfig {
	for _, item := range list {
		if item.Name == name {
			return item
		}
	}
	return nil
}

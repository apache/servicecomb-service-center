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
	"crypto/md5"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/tlsutil"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/pkg/errors"
)

// Verify Provide config verification
func Verify(c *Config) (err error) {
	if err = verifyListener(&c.Listener); err != nil {
		return
	}

	if err = verifyJoin(&c.Join); err != nil {
		return
	}

	if err = verifyTask(&c.Task); err != nil {
		return
	}

	if err = verifyRegistry(&c.Registry); err != nil {
		return
	}

	if c.Listener.TLSMount.Enabled {
		listenerTls := c.GetTLSConfig(c.Listener.TLSMount.Name)
		if listenerTls == nil {
			err = errors.Errorf("listener tls config notfound, name = %s", c.Listener.TLSMount.Name)
			return
		}
		if err = verifyTLSConfig(listenerTls); err == nil {
			return
		}
	}

	if c.Registry.TLSMount.Enabled {
		registryTls := c.GetTLSConfig(c.Listener.TLSMount.Name)
		if registryTls == nil {
			err = errors.Errorf("registry tls config notfound, name = %s", c.Registry.TLSMount.Name)
			return
		}
		if err = verifyTLSConfig(registryTls); err == nil {
			return
		}
	}

	if c.Cluster == "" {
		endpoints := strings.Split(c.Registry.Address, ",")
		sort.Strings(endpoints)
		str := strings.Join(endpoints, ",")
		c.Cluster = fmt.Sprintf("%x", md5.Sum([]byte(str)))
	}

	if c.DataDir == "" {
		c.DataDir = defaultDataDir + c.Node
	}
	c.DataDir = filepath.Dir(c.DataDir)
	return nil
}

func verifyListener(listener *Listener) (err error) {
	bindHost, bindPort, bErr := utils.SplitHostPort(listener.BindAddr, defaultBindPort)
	if bErr != nil {
		err = errors.Wrapf(bErr, "verify bind address failed, url is %s", listener.BindAddr)
		return
	}
	listener.BindAddr = bindHost + ":" + strconv.Itoa(bindPort)

	rpcHost, rpcPort, rErr := utils.SplitHostPort(listener.RPCAddr, defaultRPCPort)
	if rErr != nil {
		err = errors.Wrapf(rErr, "verify rpc address failed, url is %s", listener.RPCAddr)
		return
	}
	listener.RPCAddr = rpcHost + ":" + strconv.Itoa(rpcPort)

	peerHost, peerPort, pErr := utils.SplitHostPort(listener.PeerAddr, defaultPeerPort)
	if pErr != nil {
		err = errors.Wrapf(pErr, "verify peer address failed, url is %s", listener.PeerAddr)
		return
	}
	listener.PeerAddr = peerHost + ":" + strconv.Itoa(peerPort)
	return
}

func verifyJoin(join *Join) (err error) {
	if !join.Enabled {
		return
	}
	endpoints := strings.Split(join.Address, ",")
	for _, addr := range endpoints {
		_, _, err1 := net.SplitHostPort(addr)
		if err1 != nil {
			err = errors.Wrapf(err1, "Verify joinAddr failed, urls has %s", addr)
			return
		}
	}

	if join.RetryMax < 0 {
		join.RetryMax = defaultRetryJoinMax
	}

	if _, err1 := time.ParseDuration(join.RetryInterval); err1 != nil {
		log.Warnf("join retry interval '%s' is wrong", join.RetryInterval)
		join.RetryInterval = defaultRetryJoinInterval
	}
	return
}

func verifyTask(task *Task) (err error) {
	if task.Kind == "" {
		task.Kind = defaultTaskKind
	}

	if task.Kind == defaultTaskKind {
		for _, label := range task.Params {
			if label.Key != defaultTaskKey {
				continue
			}
			_, err1 := time.ParseDuration(label.Value)
			if err1 != nil {
				err = errors.Wrapf(err1, "Verify task params failed, key = %s, value = %s", label.Key, label.Value)
				return
			}
		}
	}
	return
}

func verifyRegistry(r *Registry) (err error) {
	endpoints := strings.Split(r.Address, ",")
	for _, addr := range endpoints {
		_, err = url.Parse(addr)
		if err != nil {
			log.Errorf(err, "Verify registry endpoints failed, urls has %s", addr)
			return err
		}
	}
	return nil
}

func verifyTLSConfig(conf *TLSConfig) (err error) {
	if conf.CAFile == "" || !utils.IsFileExist(conf.CAFile) {
		conf.CAFile = pathFromSSLEnvOrDefault(conf.Name, defaultCAName)
		if !utils.IsFileExist(conf.CAFile) {
			err = errors.Errorf("tls ca file '%s' is not found", conf.CAFile)
			return
		}
	}

	if conf.CertFile == "" || !utils.IsFileExist(conf.CertFile) {
		conf.CertFile = pathFromSSLEnvOrDefault(conf.Name, defaultCertName)
		if !utils.IsFileExist(conf.CertFile) {
			err = errors.Errorf("tls cert file '%s' is not found", conf.CertFile)
			return
		}
	}

	if conf.KeyFile == "" || !utils.IsFileExist(conf.KeyFile) {
		conf.KeyFile = pathFromSSLEnvOrDefault(conf.Name, defaultKeyName)
		if !utils.IsFileExist(conf.KeyFile) {
			err = errors.Errorf("tls key file '%s' is not found", conf.KeyFile)
			return
		}
	}

	for _, cipher := range conf.Ciphers {
		if _, ok := tlsutil.TLS_CIPHER_SUITE_MAP[cipher]; !ok {
			err = errors.Errorf("cipher %s not exist", cipher)
			return
		}
	}
	return
}

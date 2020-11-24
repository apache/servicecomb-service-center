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
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	conf := DefaultConfig()
	assert.NotNil(t, conf)
}

func TestLoadConfig(t *testing.T) {
	configFile := ""
	conf, err := LoadConfig(configFile)
	assert.Nil(t, conf)

	configFile = "./test.yaml"
	conf, err = LoadConfig(configFile)
	assert.NotNil(t, err)

	defer os.Remove(configFile)
	err = createFile(configFile, notYAMLData())
	assert.Nil(t, err)
	conf, err = LoadConfig(configFile)
	assert.NotNil(t, err)

	err = createFile(configFile, correctConfiguration())
	assert.Nil(t, err)
	conf, err = LoadConfig(configFile)
	assert.Nil(t, err)
}

func TestGetTLSConfig(t *testing.T) {
	configFile := "./test.yaml"
	defer os.Remove(configFile)
	err := createFile(configFile, correctConfiguration())
	assert.Nil(t, err)
	conf, err := LoadConfig(configFile)
	assert.Nil(t, err)

	tlsConf := conf.GetTLSConfig(conf.Listener.TLSMount.Name)
	assert.NotNil(t, tlsConf)
}

func TestMerge(t *testing.T) {
	configFile := "./test.yaml"
	defer os.Remove(configFile)
	err := createFile(configFile, correctConfiguration())
	assert.Nil(t, err)
	conf, err := LoadConfig(configFile)
	assert.Nil(t, err)

	nConf := Merge(*conf, *conf, *DefaultConfig())
	assert.NotNil(t, nConf)
}

func TestVerify(t *testing.T) {
	conf := DefaultConfig()
	conf.DataDir = ""
	err := Verify(conf)
	assert.Nil(t, err)

	configFile := "./test.yaml"
	defer os.Remove(configFile)
	err = createFile(configFile, correctConfiguration())
	assert.Nil(t, err)
	conf, err = LoadConfig(configFile)
	assert.Nil(t, err)
	err = Verify(conf)
	assert.Nil(t, err)

	bindAddr := conf.Listener.BindAddr
	conf.Listener.BindAddr = ""
	err = Verify(conf)
	conf.Listener.BindAddr = bindAddr
	assert.NotNil(t, err)

	rpcAddr := conf.Listener.RPCAddr
	conf.Listener.RPCAddr = ""
	err = Verify(conf)
	conf.Listener.RPCAddr = rpcAddr
	assert.NotNil(t, err)

	peerAddr := conf.Listener.PeerAddr
	conf.Listener.PeerAddr = ""
	err = Verify(conf)
	conf.Listener.PeerAddr = peerAddr
	assert.NotNil(t, err)

	conf.Listener.TLSMount.Enabled = true
	err = Verify(conf)
	assert.Nil(t, err)

	conf.Listener.TLSMount.Name += "_test"
	err = Verify(conf)
	conf.Listener.TLSMount.Enabled = false
	assert.NotNil(t, err)

	conf.Registry.TLSMount.Enabled = true
	err = Verify(conf)
	assert.NotNil(t, err)

	conf.Registry.TLSMount.Name += "_test"
	err = Verify(conf)
	conf.Registry.TLSMount.Enabled = false
	assert.NotNil(t, err)

	registry := conf.Registry.Address
	conf.Registry.Address = "127.0.0.1:xxx"
	err = Verify(conf)
	conf.Registry.Address = registry
	assert.NotNil(t, err)

	conf.Join.Enabled = true
	joinAddr := conf.Join.Address
	conf.Join.Address = "http://127.0.0.1:9999"
	err = Verify(conf)
	conf.Join.Address = joinAddr
	assert.NotNil(t, err)

	conf.Join.RetryMax = -1
	conf.Join.RetryInterval = "3mams"
	err = Verify(conf)
	conf.Join.Enabled = false
	assert.Nil(t, err)

	params := conf.Task.Params
	conf.Task.Kind = ""
	conf.Task.Params = []Label{{Key: "test", Value: "test"}, {Key: defaultTaskKey, Value: "3mams"}}
	err = Verify(conf)
	conf.Task.Params = params
	assert.NotNil(t, err)

	httpAddr := conf.HttpConfig.HttpAddr
	conf.HttpConfig.HttpAddr = ""
	err = Verify(conf)
	conf.HttpConfig.HttpAddr = httpAddr
	assert.NotNil(t, err)
}

func createFile(path string, data []byte) error {
	fileDir := filepath.Dir(path)
	if !utils.IsDirExist(fileDir) {
		err := os.MkdirAll(fileDir, 0640)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(path, data, 0640)
}

func notYAMLData() []byte {
	return []byte("xxxxxxxxxxxx")
}

func correctConfiguration() []byte {
	return []byte(`# run mode, supports (single, cluster)
mode: signle
# node name, must be unique on the network
node: syncer-node
# Cluster name, clustering by this name
cluster: syncer-cluster
dataDir: ./syncer-data/
listener:
  # Address used to network with other Syncers in LAN
  bindAddr: 0.0.0.0:30190
  # Address used to network with other Syncers in WAN
  advertiseAddr: ""
  # Address used to synchronize data with other Syncers
  rpcAddr: 0.0.0.0:30191
  # Address used to communicate with other cluster peers
  peerAddr: 127.0.0.1:30192
  tlsMount:
    enabled: false
    name: syncer
join:
  enabled: false
  # Address to join the network by specifying at least one existing member
  address: 127.0.0.1:30190
  # Limit the maximum of RetryJoin, default is 0, means no limit
  retryMax: 3
  retryInterval: 30s
task:
  kind: ticker
  params:
    # Time interval between timing tasks, default is 30s
    - key: interval
      value: 30s
registry:
  plugin: servicecenter
  address: http://127.0.0.1:30100
  tlsMount:
    enabled: false
    name: servicecenter
HttpConfig:
  httpAddr: 0.0.0.0:30300
  compressed: true
  compressMinBytes: 1400
tlsConfigs:
  - name: syncer
    verifyPeer: true
    minVersion: TLSv1.2
    caFile: ./certs/trust.cer
    certFile: ./certs/server.cer
    keyFile: ./certs/server_key.pem
    passphrase: ""
    ciphers:
      - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
      - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
      - TLS_RSA_WITH_AES_256_GCM_SHA384
      - TLS_RSA_WITH_AES_128_GCM_SHA256
  - name: servicecenter
    verifyPeer: false
    caFile: ./certs/trust.cer
    certFile: ./certs/server.cer
    keyFile: ./certs/server_key.pem
`)
}

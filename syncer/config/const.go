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

const (
	defaultBindPort          = 30190
	defaultRPCPort           = 30191
	defaultPeerPort          = 30192
	defaultTaskKind          = "ticker"
	defaultTaskKey           = "interval"
	defaultTaskValue         = "30s"
	defaultDataDir           = "./syncer-data/"
	defaultDCPluginName      = "servicecenter"
	defaultRetryJoinMax      = 3
	defaultRetryJoinInterval = "30s"

	defaultEnvSSLRoot = "SSL_ROOT"
	defaultCertsDir   = "certs"
	defaultCAName     = "trust.cer"
	defaultCertName   = "server.cer"
	defaultKeyName    = "server_key.pem"
	// ModeSingle run as a single server
	ModeSingle = "single"
	// ModeCluster run as a cluster peer
	ModeCluster = "cluster"
)

// Config is the configuration that can be set for Syncer. Some of these
// configurations are exposed as command-line flags.
type Config struct {
	Mode       string       `yaml:"mode"`
	Node       string       `yaml:"node"`
	Cluster    string       `yaml:"cluster"`
	DataDir    string       `yaml:"dataDir"`
	Listener   Listener     `yaml:"listener"`
	Join       Join         `yaml:"join"`
	Task       Task         `yaml:"task"`
	Registry   Registry     `yaml:"registry"`
	TLSConfigs []*TLSConfig `yaml:"tlsConfigs"`
}

// Listener Configuration for Syncer listener
type Listener struct {
	BindAddr      string `yaml:"bindAddr"`
	AdvertiseAddr string `yaml:"advertiseAddr"`
	RPCAddr       string `yaml:"rpcAddr"`
	PeerAddr      string `yaml:"peerAddr"`
	TLSMount      Mount  `yaml:"tlsMount"`
}

// Join Configuration for Syncer join the network
type Join struct {
	Enabled       bool   `yaml:"enabled"`
	Address       string `yaml:"address"`
	RetryMax      int    `yaml:"retryMax"`
	RetryInterval string `yaml:"retryInterval"`
}

// Task
type Task struct {
	Kind   string  `yaml:"kind"`
	Params []Label `yaml:"params"`
}

// Label pair of key and value
type Label struct {
	Key   string
	Value string
}

// Registry configuration
type Registry struct {
	// Address is the service registry address.
	Address  string `json:"address"`
	Plugin   string `yaml:"plugin"`
	TLSMount Mount  `yaml:"tlsMount"`
}

// Mount Specifying config and purpose
type Mount struct {
	Enabled bool   `yaml:"enabled"`
	Name    string `yaml:"name"`
}

// TLSConfig tls configuration
type TLSConfig struct {
	Name       string   `yaml:"name"`
	VerifyPeer bool     `yaml:"verifyPeer"`
	MinVersion string   `yaml:"minVersion"`
	Passphrase string   `yaml:"passphrase"`
	CAFile     string   `yaml:"caFile"`
	CertFile   string   `yaml:"certFile"`
	KeyFile    string   `yaml:"keyFile"`
	Ciphers    []string `yaml:"ciphers"`
}

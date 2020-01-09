// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package diagnose

import (
	"github.com/apache/servicecomb-service-center/pkg/client/etcd"
	"github.com/apache/servicecomb-service-center/pkg/util"
	root "github.com/apache/servicecomb-service-center/scctl/pkg/cmd"
	"github.com/spf13/cobra"
	"path/filepath"
)

var EtcdClientConfig etcd.Config

func init() {
	root.RootCmd().AddCommand(NewDiagnoseCommand(root.RootCmd()))
}

func NewDiagnoseCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "diagnose [options]",
		Short:   "Output the service center diagnostic report",
		Run:     DiagnoseCommandFunc,
		Example: parent.CommandPath() + ` diagnose --addr "http://127.0.0.1:30100" --etcd-addr "http://127.0.0.1:2379";`,
	}

	cmd.Flags().StringVar(&EtcdClientConfig.Addrs, "etcd-addr",
		util.GetEnvString("CSE_REGISTRY_ADDRESS", "http://127.0.0.1:2379"),
		"the http addr and port of etcd endpoints")
	cmd.Flags().StringVar(&EtcdClientConfig.CertFile, "etcd-cert",
		filepath.Join(util.GetEnvString("SSL_ROOT", "."), "server.cer"),
		"the certificate file path to access etcd, can be overrode by env $SSL_ROOT/server.cer.")
	cmd.Flags().StringVar(&EtcdClientConfig.CertKeyFile, "etcd-key",
		filepath.Join(util.GetEnvString("SSL_ROOT", "."), "server_key.pem"),
		"the key file path to access etcd, can be overrode by env $SSL_ROOT/server_key.pem.")
	cmd.Flags().StringVar(&EtcdClientConfig.CAFile, "etcd-ca",
		filepath.Join(util.GetEnvString("SSL_ROOT", "."), "trust.cer"),
		"the CA file path  to access etcd, can be overrode by env $SSL_ROOT/trust.cer.")
	cmd.Flags().StringVar(&EtcdClientConfig.CertKeyPWDPath, "etcd-pass-file",
		filepath.Join(util.GetEnvString("SSL_ROOT", "."), "cert_pwd"),
		"the passphase file path to decrypt key file, can be overrode by env $SSL_ROOT/cert_pwd.")
	cmd.Flags().StringVar(&EtcdClientConfig.CertKeyPWD, "etcd-pass", "",
		"the passphase string to decrypt key file.")

	return cmd
}

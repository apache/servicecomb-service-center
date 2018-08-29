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

package cmd

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/client/sc"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/scctl/pkg/version"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	ExitSuccess = iota
	ExitError
)

var rootCmd = &cobra.Command{
	Use:   version.TOOL_NAME + " <command>",
	Short: "The admin control command of service center",
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "make the operation more talkative")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("verbose"); v {
			os.Setenv("DEBUG_MODE", "1")
		}
	}

	rootCmd.PersistentFlags().StringVar(&sc.Addr, "addr",
		"http://"+util.GetEnvString("HOSTING_SERVER_IP", "127.0.0.1")+":30100",
		"the http host and port of service center, can be overrode by env HOSTING_SERVER_IP.")

	rootCmd.PersistentFlags().StringVarP(&sc.Token, "token", "t", "",
		"the auth token string to access service center.")

	rootCmd.PersistentFlags().BoolVarP(&sc.VerifyPeer, "peer", "p", false,
		"verify service center certificates.")
	rootCmd.PersistentFlags().StringVar(&sc.CertPath, "cert",
		filepath.Join(util.GetEnvString("SSL_ROOT", "."), "server.cer"),
		"the certificate file path to access service center, can be overrode by $SSL_ROOT/server.cer.")
	rootCmd.PersistentFlags().StringVar(&sc.KeyPath, "key",
		filepath.Join(util.GetEnvString("SSL_ROOT", "."), "server_key.pem"),
		"the key file path to access service center, can be overrode by $SSL_ROOT/server_key.pem.")
	rootCmd.PersistentFlags().StringVar(&sc.CAPath, "ca",
		filepath.Join(util.GetEnvString("SSL_ROOT", "."), "trust.cer"),
		"the CA file path  to access service center, can be overrode by $SSL_ROOT/trust.cer.")
	rootCmd.PersistentFlags().StringVar(&sc.KeyPassPath, "pass-file",
		filepath.Join(util.GetEnvString("SSL_ROOT", "."), "cert_pwd"),
		"the passphase file path to decrypt key file, can be overrode by $SSL_ROOT/cert_pwd.")
	rootCmd.PersistentFlags().StringVar(&sc.KeyPass, "pass", "",
		"the passphase string to decrypt key file.")
}

func RootCmd() *cobra.Command {
	return rootCmd
}

func StopAndExit(code int, args ...interface{}) {
	if len(args) == 0 {
		os.Exit(code)
	}

	if code == ExitSuccess {
		fmt.Fprintln(os.Stdout, args...)
	} else {
		fmt.Fprintln(os.Stderr, args...)
	}
	os.Exit(code)
}

func Run() {
	RootCmd().SetUsageFunc(UsageFunc)
	// Show usage in help command
	RootCmd().SetHelpTemplate(`{{.UsageString}}`)

	err := RootCmd().Execute()
	if err != nil {
		StopAndExit(ExitError, err)
	}
	StopAndExit(ExitSuccess)
}

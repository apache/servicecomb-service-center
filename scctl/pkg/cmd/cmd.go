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
	"github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/scctl/pkg/version"
	"github.com/spf13/cobra"
	"os"
	"time"
)

const (
	ExitSuccess = iota
	ExitError
)

var rootCmd = &cobra.Command{
	Use:   version.TOOL_NAME + " <command>",
	Short: "The admin control command of service center",
}
var ScClientConfig client.Config

func init() {
	var timeout string
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "make the operation more talkative")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("verbose"); v {
			if err := os.Setenv("DEBUG_MODE", "1"); err != nil {
				StopAndExit(ExitError, err)
			}
		}
		if d, err := time.ParseDuration(timeout); err == nil && d > 0 {
			ScClientConfig.RequestTimeout = d
		}
	}

	rootCmd.PersistentFlags().StringSliceVar(&ScClientConfig.Endpoints, "addr",
		[]string{"http://" + util.GetEnvString("HOSTING_SERVER_IP", "127.0.0.1") + ":30100"},
		"the http host and port of service center, can be overrode by env HOSTING_SERVER_IP.")

	rootCmd.PersistentFlags().StringVar(&ScClientConfig.Token, "token", "",
		"the auth token string to access service center.")

	rootCmd.PersistentFlags().BoolVarP(&ScClientConfig.VerifyPeer, "peer", "p", false,
		"verify service center certificates.")
	rootCmd.PersistentFlags().StringVar(&ScClientConfig.CertFile, "cert", "",
		"the certificate file path to access service center.")
	rootCmd.PersistentFlags().StringVar(&ScClientConfig.CertKeyFile, "key", "",
		"the key file path to access service center.")
	rootCmd.PersistentFlags().StringVar(&ScClientConfig.CAFile, "ca", "",
		"the CA file path  to access service center.")
	rootCmd.PersistentFlags().StringVar(&ScClientConfig.CertKeyPWDPath, "pass-file", "",
		"the passphase file path to decrypt key file.")
	rootCmd.PersistentFlags().StringVar(&ScClientConfig.CertKeyPWD, "pass", "",
		"the passphase string to decrypt key file.")

	rootCmd.PersistentFlags().StringVarP(&timeout, "timeout", "t", "10s",
		"the maximum time allowed for the request.")
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

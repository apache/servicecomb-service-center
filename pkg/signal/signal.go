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

package signal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

func RegisterListener() {
	go HandleSignals()
}

func HandleSignals() {
	defer log.Flush()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	wait := 5 * time.Second
	for sig := range sigCh {
		switch sig {
		case syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM:
			<-time.After(wait)
			log.Warn(fmt.Sprintf("waiting for server response timed out(%s), force shutdown", wait))
			os.Exit(1)
		default:
			log.Warn(fmt.Sprintf("received signal '%v'", sig))
		}
	}
}

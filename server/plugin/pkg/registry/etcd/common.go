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

package etcd

import (
	"math"
	"time"
)

const (
	// here will new an etcd connection after about 30s(=5s * 3 + (backoff:8s))
	// when the connected etcd member was hung but tcp is still alive
	healthCheckTimeout    = 5 * time.Second
	healthCheckRetryTimes = 3

	// see google.golang.org/grpc/keepalive/keepalive.go
	// after a duration of this time if the client doesn't see any activity
	// it pings the server to see if the transport is still alive.
	keepAliveTime    = 2 * time.Second
	keepAliveTimeout = 5 * time.Second

	// see github.com/coreos/etcd/clientv3/options.go
	maxSendMsgSize = 10 * 1024 * 1024 // 10MB
	maxRecvMsgSize = math.MaxInt32
)

const (
	OperationCompact     = "COMPACT"
	OperationTxn         = "TXN"
	OperationLeaseGrant  = "LEASE_GRANT"
	OperationLeaseRenew  = "LEASE_RENEW"
	OperationLeaseRevoke = "LEASE_REVOKE"
	OperationSyncMembers = "SYNC"
)

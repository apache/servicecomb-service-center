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

package plugin

import (
	"github.com/apache/servicecomb-service-center/server/plugin/auditlog"
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	"github.com/apache/servicecomb-service-center/server/plugin/tls"
	"github.com/apache/servicecomb-service-center/server/plugin/tracing"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	"github.com/go-chassis/cari/security"
)

const (
	UUID Name = iota
	AuditLog
	AUTH
	CIPHER
	QUOTA
	REGISTRY
	TRACING
	TLS
	DISCOVERY
	typeEnd // for counting
)

var pluginNames = map[Name]string{
	UUID:      "uuid",
	AuditLog:  "auditlog",
	AUTH:      "auth",
	CIPHER:    "cipher",
	QUOTA:     "quota",
	REGISTRY:  "registry",
	TRACING:   "trace",
	DISCOVERY: "discovery",
	TLS:       "ssl",
}

func (pm *Manager) Discovery() discovery.AdaptorRepository {
	return pm.Instance(DISCOVERY).(discovery.AdaptorRepository)
}
func (pm *Manager) Registry() registry.Registry {
	return pm.Instance(REGISTRY).(registry.Registry)
}
func (pm *Manager) UUID() uuid.UUID { return pm.Instance(UUID).(uuid.UUID) }
func (pm *Manager) AuditLog() auditlog.AuditLogger {
	return pm.Instance(AuditLog).(auditlog.AuditLogger)
}
func (pm *Manager) Auth() auth.Auth              { return pm.Instance(AUTH).(auth.Auth) }
func (pm *Manager) Cipher() security.Cipher      { return pm.Instance(CIPHER).(security.Cipher) }
func (pm *Manager) Quota() quota.Manager         { return pm.Instance(QUOTA).(quota.Manager) }
func (pm *Manager) Tracing() (v tracing.Tracing) { return pm.Instance(TRACING).(tracing.Tracing) }
func (pm *Manager) TLS() tls.TLS                 { return pm.Instance(TLS).(tls.TLS) }

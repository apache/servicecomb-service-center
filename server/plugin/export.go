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
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/auditlog"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/auth"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/quota"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/tls"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/tracing"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/uuid"
	"github.com/go-chassis/foundation/security"
)

const (
	UUID PluginName = iota
	AUDIT_LOG
	AUTH
	CIPHER
	QUOTA
	REGISTRY
	TRACING
	TLS
	DISCOVERY
	typeEnd // for counting
)

var pluginNames = map[PluginName]string{
	UUID:      "uuid",
	AUDIT_LOG: "auditlog",
	AUTH:      "auth",
	CIPHER:    "cipher",
	QUOTA:     "quota",
	REGISTRY:  "registry",
	TRACING:   "trace",
	DISCOVERY: "discovery",
	TLS:       "ssl",
}

func (pm *PluginManager) Discovery() discovery.AdaptorRepository {
	return pm.Instance(DISCOVERY).(discovery.AdaptorRepository)
}
func (pm *PluginManager) Registry() registry.Registry {
	return pm.Instance(REGISTRY).(registry.Registry)
}
func (pm *PluginManager) UUID() uuid.UUID { return pm.Instance(UUID).(uuid.UUID) }
func (pm *PluginManager) AuditLog() auditlog.AuditLogger {
	return pm.Instance(AUDIT_LOG).(auditlog.AuditLogger)
}
func (pm *PluginManager) Auth() auth.Auth              { return pm.Instance(AUTH).(auth.Auth) }
func (pm *PluginManager) Cipher() security.Cipher      { return pm.Instance(CIPHER).(security.Cipher) }
func (pm *PluginManager) Quota() quota.QuotaManager    { return pm.Instance(QUOTA).(quota.QuotaManager) }
func (pm *PluginManager) Tracing() (v tracing.Tracing) { return pm.Instance(TRACING).(tracing.Tracing) }
func (pm *PluginManager) TLS() tls.TLS                 { return pm.Instance(TLS).(tls.TLS) }

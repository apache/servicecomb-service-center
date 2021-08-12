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

package state

import (
	"errors"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

var plugins = make(map[kvstore.Type]Plugin)

func Plugins() map[kvstore.Type]Plugin {
	return plugins
}

func Register(p Plugin) (id kvstore.Type, err error) {
	if p == nil || len(p.Name()) == 0 || p.Config() == nil {
		return kvstore.TypeError, errors.New("invalid parameter")
	}

	id, err = kvstore.RegisterType(p.Name())
	if err != nil {
		return
	}

	plugins[id] = p

	log.Info(fmt.Sprintf("install new type %d:%s->%s", id, p.Name(), p.Config().Key))
	return
}

func MustRegister(name string, prefix string, opts ...Option) kvstore.Type {
	opts = append(opts, WithPrefix(prefix))
	options := ToOptions(opts...)

	id, err := Register(NewPlugin(name, &options))
	if err != nil {
		panic(err)
	}
	return id
}

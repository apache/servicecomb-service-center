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

package plugin

import (
	"github.com/bytedance/sonic"
	"github.com/go-chassis/cari/codec"
	codecChassis "github.com/go-chassis/go-chassis/v2/pkg/codec"
)

func init() {
	codecChassis.Install("bytedance/sonic", newDefault)
}

type Sonic struct {
}

func newDefault(opts codecChassis.Options) (codec.Codec, error) {
	return &Sonic{}, nil
}
func (s *Sonic) Encode(v any) ([]byte, error) {
	return sonic.Marshal(v)
}

func (s *Sonic) Decode(data []byte, v any) error {
	return sonic.Unmarshal(data, v)
}

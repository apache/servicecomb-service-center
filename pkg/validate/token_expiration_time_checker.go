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

package validate

import (
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

// TokenExpirationTimeChecker validates Account.TokenExpirationTime should >= 15m and <=24h
type TokenExpirationTimeChecker struct {
}

const (
	//MinTokenDuration the minimum time duration of a token
	MinTokenDuration time.Duration = time.Minute * 15
	//MaxTokenDuration the maximum time duration of a token
	MaxTokenDuration time.Duration = time.Hour * 24
)

//MatchString ensures TokenExpirationTime is a valid time.Duration and in legal range.
func (p *TokenExpirationTimeChecker) MatchString(s string) bool {
	duration, err := time.ParseDuration(s)
	if err != nil {
		log.Error(fmt.Sprintf("Invalid duration value '%s'", s), err)
		return false
	}
	if duration > MaxTokenDuration || duration < MinTokenDuration {
		log.Error(fmt.Sprintf("TokenExpirationTime('%s') should >= 15m and <= 24h.", s), err)
		return false
	}
	return true
}

func (p *TokenExpirationTimeChecker) String() string {
	return "TokenExpirationTime"
}

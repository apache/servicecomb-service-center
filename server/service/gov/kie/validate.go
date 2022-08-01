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

package kie

import (
	"fmt"
)

type Validator struct {
}

type ErrIllegalItem struct {
	err string
	val interface{}
}

var (
	methodSet map[string]bool
)

func (e *ErrIllegalItem) Error() string {
	return fmt.Sprintf("illegal item : %v , msg: %s", e.val, e.err)
}

func (d *Validator) Validate(kind string, spec interface{}) error {
	switch kind {
	case "match-group":
		return matchValidate(spec)
	case "retry":
		return retryValidate(spec)
	case "rate-limiting":
		return rateLimitingValidate(spec)
	case "circuit-breaker":
	case "instance-isolation":
	case "fault-injection":
	case "bulkhead":
	case "loadbalancer":
		return nil
	default:
		return &ErrIllegalItem{"not support kind yet", kind}
	}
	return nil
}

func matchValidate(val interface{}) error {
	spec, ok := val.(map[string]interface{})
	if !ok {
		return &ErrIllegalItem{"can not cast to map", val}
	}
	if spec[Matches] == nil {
		return nil
	}
	alias, ok := spec[Alias].(string)
	if !ok {
		return &ErrIllegalItem{"alias must be string", alias}
	}
	matches, ok := spec[Matches].([]interface{})
	if !ok {
		return &ErrIllegalItem{"don't have matches", spec}
	}
	for _, match := range matches {
		match, ok := match.(map[string]interface{})
		if !ok {
			return &ErrIllegalItem{"match can not cast to map", match}
		}
		if match["name"] == nil {
			return &ErrIllegalItem{"match's name can not be null", match}
		}
		if match["apiPath"] == nil && match["headers"] == nil && match[Method] == nil {
			return &ErrIllegalItem{"match must have a match item [apiPath/headers/methods]", match}
		}
		//apiPath & headers do not check
		if match[Method] != nil {
			methods, ok := match[Method].([]interface{})
			if !ok {
				return &ErrIllegalItem{"methods must be a list", match}
			}
			for _, method := range methods {
				methodStr, ok := method.(string)
				if !ok {
					return &ErrIllegalItem{"method must be a string", method}
				}
				if !methodSet[methodStr] {
					return &ErrIllegalItem{"method must be one of the GET/POST/PUT/DELETE/PATCH", method}
				}
			}
		}
	}
	return nil
}

func retryValidate(val interface{}) error {
	err := policyValidate(val)
	if err != nil {
		return err
	}
	return nil
}

func rateLimitingValidate(val interface{}) error {
	err := policyValidate(val)
	if err != nil {
		return err
	}
	return nil
}

func policyValidate(val interface{}) error {
	spec, ok := val.(map[string]interface{})
	if !ok {
		return &ErrIllegalItem{"policy can not cast to map", val}
	}
	if spec[Rules] != nil {
		rules, ok := spec[Rules].(map[string]interface{})
		if !ok {
			return &ErrIllegalItem{"policy's rules can not cast to map", spec}
		}
		if "" == rules["match"] {
			return &ErrIllegalItem{"policy's rules match can not be nil", spec}
		}
	}
	return nil
}

func init() {
	methodSet = make(map[string]bool)
	methodSet["GET"] = true
	methodSet["POST"] = true
	methodSet["DELETE"] = true
	methodSet["PUT"] = true
	methodSet["PATCH"] = true
}

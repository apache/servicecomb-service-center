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

package validate_test

import (
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/validate"
	"github.com/stretchr/testify/assert"
)

func TestNewVersionRegexp(t *testing.T) {
	log.Info("normal")

	t.Run("latest", func(t *testing.T) {
		vr := validate.NewVersionRegexp(false)
		assert.Equal(t, false, vr.MatchString("latest"))
		vr = validate.NewVersionRegexp(true)
		assert.Equal(t, true, vr.MatchString("latest"))
	})

	t.Run("range", func(t *testing.T) {
		vr := validate.NewVersionRegexp(false)
		assert.Equal(t, false, vr.MatchString("1.1-2.2"))
		vr = validate.NewVersionRegexp(true)
		assert.Equal(t, false, vr.MatchString("-"))
		assert.Equal(t, false, vr.MatchString("1.1-"))
		assert.Equal(t, false, vr.MatchString("-1.1"))
		assert.Equal(t, false, vr.MatchString("1.a-2.b"))
		assert.Equal(t, false, vr.MatchString("1.-.2"))
		assert.Equal(t, false, vr.MatchString("60000-1"))
		assert.Equal(t, true, vr.MatchString("1.1-2.2"))
		assert.Equal(t, true, vr.MatchString("1.1.1.1-2.2.2.2"))
		assert.Equal(t, false, vr.MatchString("1.1.1.1.1-2.2.2.2"))
		assert.Equal(t, false, vr.MatchString("1.1.1.1-2.2.2.2.2"))
	})

	t.Run("atLess", func(t *testing.T) {
		vr := validate.NewVersionRegexp(false)
		assert.Equal(t, false, vr.MatchString("1.0+"))
		vr = validate.NewVersionRegexp(true)
		assert.Equal(t, false, vr.MatchString("+"))
		assert.Equal(t, false, vr.MatchString("+1.0"))
		assert.Equal(t, false, vr.MatchString("1.a+"))
		assert.Equal(t, false, vr.MatchString(".1+"))
		assert.Equal(t, false, vr.MatchString("1.+"))
		assert.Equal(t, false, vr.MatchString(".+"))
		assert.Equal(t, false, vr.MatchString("60000+"))
		assert.Equal(t, true, vr.MatchString("1.0+"))
		assert.Equal(t, true, vr.MatchString("1.0.0.0+"))
		assert.Equal(t, false, vr.MatchString("1.0.0.0.0+"))
	})

	t.Run("explicit", func(t *testing.T) {
		vr := validate.NewVersionRegexp(false)
		assert.Equal(t, false, vr.MatchString(""))
		assert.Equal(t, false, vr.MatchString("a"))
		assert.Equal(t, false, vr.MatchString("60000"))
		assert.Equal(t, false, vr.MatchString("."))
		assert.Equal(t, false, vr.MatchString("1."))
		assert.Equal(t, false, vr.MatchString(".1"))
		assert.Equal(t, true, vr.MatchString("1.4"))
		vr = validate.NewVersionRegexp(true)
		assert.Equal(t, false, vr.MatchString(""))
		assert.Equal(t, false, vr.MatchString("a"))
		assert.Equal(t, false, vr.MatchString("60000"))
		assert.Equal(t, false, vr.MatchString("."))
		assert.Equal(t, false, vr.MatchString("1."))
		assert.Equal(t, false, vr.MatchString(".1"))
		assert.Equal(t, true, vr.MatchString("1.4"))
		assert.Equal(t, true, vr.MatchString("1.4.0.0"))
		assert.Equal(t, false, vr.MatchString("1.4.0.0.0"))
	})

	log.Info("exception")

	t.Run("MatchString & String", func(t *testing.T) {
		vr := validate.VersionRegexp{}
		assert.Equal(t, true, vr.MatchString(""))
		assert.NotEqual(t, "", vr.String())
		vr = validate.VersionRegexp{Fuzzy: true}
		assert.Equal(t, true, vr.MatchString(""))
		assert.NotEqual(t, "", vr.String())
	})
}

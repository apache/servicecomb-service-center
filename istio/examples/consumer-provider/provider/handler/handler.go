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

package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	_ "github.com/go-chassis/go-chassis/v2/bootstrap"
	rf "github.com/go-chassis/go-chassis/v2/server/restful"
)

type Greating struct{}

//URLPatterns helps to respond for corresponding API calls
func (r *Greating) URLPatterns() []rf.Route {
	return []rf.Route{
		{Method: http.MethodGet, Path: "/sqrt", ResourceFunc: r.Sqrt,
			Returns: []*rf.Returns{{Code: 200}}},
	}
}

func (r *Greating) Sqrt(ctx *rf.Context) {
	xstr := ctx.ReadQueryParameter("x")
	x, err := strconv.Atoi(xstr)
	if err != nil {
		return
	}
	if x < 1 {
		ctx.Write([]byte(fmt.Sprintf("Square root of %d is %f", x, math.Sqrt(float64(x)))))
		return
	}
	var sum float64
	for i := 0; i < x; i++ {
		sum += math.Sqrt(float64(i))
	}
	ctx.Write([]byte(fmt.Sprintf("Sum of square root from 1 to %d is %f", x, sum)))
}

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
package buildin

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
	"github.com/openzipkin/zipkin-go-opentracing/types"
	"strconv"
)

type Span struct {
	TraceID string `thrift:"trace_id,1" db:"trace_id" json:"traceId"`
	// unused field # 2
	Name        string        `thrift:"name,3" db:"name" json:"name"`
	ID          string        `thrift:"id,4" db:"id" json:"id"`
	ParentID    string        `thrift:"parent_id,5" db:"parent_id" json:"parentId,omitempty"`
	Annotations []*Annotation `thrift:"annotations,6" db:"annotations" json:"annotations"`
	// unused field # 7
	BinaryAnnotations []*BinaryAnnotation `thrift:"binary_annotations,8" db:"binary_annotations" json:"binaryAnnotations"`
	//Debug bool `thrift:"debug,9" db:"debug" json:"debug,omitempty"`
	Timestamp *int64 `thrift:"timestamp,10" db:"timestamp" json:"timestamp,omitempty"`
	Duration  *int64 `thrift:"duration,11" db:"duration" json:"duration,omitempty"`
	//TraceIDHigh *int64 `thrift:"trace_id_high,12" db:"trace_id_high" json:"trace_id_high,omitempty"`
}

type Annotation struct {
	Timestamp int64     `thrift:"timestamp,1" db:"timestamp" json:"timestamp"`
	Value     string    `thrift:"value,2" db:"value" json:"value"`
	Host      *Endpoint `thrift:"host,3" db:"host" json:"host,omitempty"`
}

type BinaryAnnotation struct {
	Key   string `thrift:"key,1" db:"key" json:"key"`
	Value string `thrift:"value,2" db:"value" json:"value"`
	//AnnotationType AnnotationType `thrift:"annotation_type,3" db:"annotation_type" json:"annotation_type"`
	//Host *Endpoint `thrift:"host,4" db:"host" json:"host,omitempty"`
}

type Endpoint struct {
	Ipv4        string `thrift:"ipv4,1" db:"ipv4" json:"ipv4"`
	Port        int16  `thrift:"port,2" db:"port" json:"port"`
	ServiceName string `thrift:"service_name,3" db:"service_name" json:"serviceName"`
	Ipv6        []byte `thrift:"ipv6,4" db:"ipv6" json:"ipv6,omitempty"`
}

func (s *Span) FromZipkinSpan(span *zipkincore.Span) {
	traceId := new(types.TraceID)
	traceId.Low = uint64(span.TraceID)
	traceId.High = uint64(*(span.TraceIDHigh))
	s.TraceID = traceId.ToHex()
	s.Duration = span.Duration

	s.ID = strconv.FormatUint(uint64(span.ID), 16)
	if span.ParentID != nil {
		s.ParentID = strconv.FormatUint(uint64(*(span.ParentID)), 16)
	}

	s.Name = span.Name
	s.Timestamp = span.Timestamp

	for _, a := range span.Annotations {
		s.Annotations = append(s.Annotations, &Annotation{
			Timestamp: a.Timestamp,
			Value:     a.Value,
			Host: &Endpoint{
				Ipv4:        util.InetNtoa(uint32(a.Host.Ipv4)),
				Port:        a.Host.Port,
				ServiceName: a.Host.ServiceName,
				Ipv6:        a.Host.Ipv6,
			},
		})
	}

	for _, ba := range span.BinaryAnnotations {
		if zipkincore.SERVER_ADDR == ba.Key {
			continue
		}
		s.BinaryAnnotations = append(s.BinaryAnnotations, &BinaryAnnotation{
			Key:   ba.Key,
			Value: string(ba.Value),
		})
	}
}

func FromZipkinSpan(span *zipkincore.Span) (s Span) {
	s.FromZipkinSpan(span)
	return
}

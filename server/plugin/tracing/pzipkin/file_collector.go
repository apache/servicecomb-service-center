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

package pzipkin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/go-chassis/foundation/gopool"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type FileCollector struct {
	// Timeout is the timeout of sending span to chan
	Timeout time.Duration
	// Interval is interval to log
	Interval time.Duration
	// BatchSize is the log batch size
	BatchSize int
	logger    *lumberjack.Logger
	c         chan *zipkincore.Span
}

func (f *FileCollector) Collect(span *zipkincore.Span) error {
	timer := time.NewTimer(f.Timeout)
	select {
	case f.c <- span:
		timer.Stop()
	case <-timer.C:
		log.Error(fmt.Sprintf("send span to handle channel timed out(%s)", f.Timeout), nil)
	}
	return nil
}

func (f *FileCollector) Close() error {
	return f.logger.Close()
}

func (f *FileCollector) write(batch []*zipkincore.Span) (c int) {
	if len(batch) == 0 {
		return
	}

	newLine := [...]byte{'\n'}
	for _, span := range batch {
		s := FromZipkinSpan(span)
		b, err := json.Marshal(s)
		if err != nil {
			log.Error("marshal span failed", err)
			continue
		}
		_, err = f.logger.Write(b)
		if err != nil {
			log.Error("", err)
		}
		_, err = f.logger.Write(newLine[:])
		if err != nil {
			log.Error("", err)
		}
		c++
	}
	return
}

func (f *FileCollector) Run() {
	gopool.Go(func(ctx context.Context) {
		var (
			batch []*zipkincore.Span
			prev  []*zipkincore.Span
			t     = time.NewTicker(f.Interval)
			max   = f.BatchSize * 2
		)
		for {
			select {
			case <-ctx.Done():
				f.write(batch)
				return
			case span := <-f.c:
				l := len(batch)
				if l >= max {
					dispose := l - f.BatchSize
					log.Error(fmt.Sprintf("backlog is full, dispose %d span(s), max: %d",
						dispose, max), nil)
					batch = batch[dispose:] // allocate more
				}

				batch = append(batch, span)

				l = len(batch)
				if l < f.BatchSize {
					continue
				}

				if c := f.write(batch); c == 0 {
					continue
				}

				if prev != nil {
					batch, prev = prev[:0], batch
				} else {
					prev, batch = batch, batch[len(batch):] // new one
				}
			case <-t.C:
				if c := f.write(batch); c > 0 {
					batch = batch[:0]
				}
			}
		}
	})
}

func NewFileCollector(path string) (*FileCollector, error) {
	fc := &FileCollector{
		Timeout:   5 * time.Second,
		Interval:  10 * time.Second,
		BatchSize: 100,
		logger: &lumberjack.Logger{
			Filename:   path,
			MaxSize:    int(config.GetLog().LogRotateSize), // megabytes
			MaxBackups: int(config.GetLog().LogBackupCount),
			LocalTime:  true,
			Compress:   true,
		},
		c: make(chan *zipkincore.Span, 1000),
	}
	fc.Run()
	return fc, nil
}

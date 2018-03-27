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
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
	"os"
	"time"
)

type FileCollector struct {
	Fd        *os.File
	Interval  time.Duration
	BatchSize int
	c         chan *zipkincore.Span
}

func (f *FileCollector) Collect(span *zipkincore.Span) error {
	if f.Fd == nil {
		return fmt.Errorf("required FD to write")
	}

	f.c <- span
	return nil
}

func (f *FileCollector) Close() error {
	return f.Fd.Close()
}

func (f *FileCollector) write(batch []*zipkincore.Span) (c int) {
	if len(batch) == 0 {
		return
	}

	if err := f.checkFile(); err != nil {
		util.Logger().Errorf(err, "check tracing file failed")
		return
	}

	newLine := [...]byte{'\n'}
	w := bufio.NewWriter(f.Fd)
	for _, span := range batch {
		s := FromZipkinSpan(span)
		b, err := json.Marshal(s)
		if err != nil {
			util.Logger().Errorf(err, "marshal span failed")
			continue
		}
		w.Write(b)
		w.Write(newLine[:])
		c++
	}
	if err := w.Flush(); err != nil {
		util.Logger().Errorf(err, "write span to file failed")
	}
	return
}

func (f *FileCollector) checkFile() error {
	if util.PathExist(f.Fd.Name()) {
		return nil
	}

	stat, err := f.Fd.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %s", f.Fd.Name(), err)
	}

	util.Logger().Warnf(nil, "tracing file %s does not exist, re-create one", f.Fd.Name())
	fd, err := os.OpenFile(f.Fd.Name(), os.O_APPEND|os.O_CREATE|os.O_RDWR, stat.Mode())
	if err != nil {
		return fmt.Errorf("open %s: %s", f.Fd.Name(), err)
	}

	var old *os.File
	f.Fd, old = fd, f.Fd
	if err := old.Close(); err != nil {
		util.Logger().Errorf(err, "close %s", f.Fd.Name())
	}
	return nil
}

func (f *FileCollector) Run(stopCh <-chan struct{}) {
	var (
		batch []*zipkincore.Span
		prev  []*zipkincore.Span
		i     = f.Interval * 10
		t     = time.NewTicker(f.Interval)
		nr    = time.Now().Add(i)
		max   = f.BatchSize * 2
	)
	for {
		select {
		case <-stopCh:
			f.write(batch)
			return
		case span := <-f.c:
			batch = append(batch, span)
			if len(batch) >= f.BatchSize {
				if len(batch) > max {
					dispose := len(batch) - f.BatchSize
					util.Logger().Errorf(nil, "backlog is full, dispose %d span(s), max: %d",
						dispose, max)
					batch = batch[dispose:] // allocate more
				}
				if c := f.write(batch); c == 0 {
					continue
				}
				if prev != nil {
					batch, prev = prev[:0], batch
				} else {
					prev, batch = batch, batch[len(batch):] // new one
				}
			}
		case <-t.C:
			if time.Now().After(nr) {
				util.LogRotateFile(f.Fd.Name(),
					int(core.ServerInfo.Config.LogRotateSize),
					int(core.ServerInfo.Config.LogBackupCount),
				)
				nr = time.Now().Add(i)
			}

			if c := f.write(batch); c > 0 {
				batch = batch[:0]
			}
		}
	}
}

func NewFileCollector(path string) (*FileCollector, error) {
	fd, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	fc := &FileCollector{
		Fd:        fd,
		Interval:  10 * time.Second,
		BatchSize: 100,
		c:         make(chan *zipkincore.Span, 1000),
	}
	util.Go(fc.Run)
	return fc, nil
}

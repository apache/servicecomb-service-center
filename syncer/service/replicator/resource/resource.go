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

package resource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"

	"github.com/apache/servicecomb-service-center/pkg/log"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	"github.com/go-chassis/cari/sync"
)

const (
	Success int32 = iota
	Fail
	Skip
	MicroNonExist
	InstNonExist
	NonImplement
)

const (
	ResultStatusSkip          = "skip"
	ResultStatusSuccess       = "success"
	ResultStatusFail          = "fail"
	ResultStatusMicroNonExist = "microNonExist"
	ResultStatusInstNonExist  = "instNonExist"
	ResultStatusNonImplement  = "nonImplement"
	oneDaySecond              = 86400
)

var codeDescriber = map[int32]string{
	Skip:          ResultStatusSkip,
	Success:       ResultStatusSuccess,
	Fail:          ResultStatusFail,
	MicroNonExist: ResultStatusMicroNonExist,
	InstNonExist:  ResultStatusInstNonExist,
	NonImplement:  ResultStatusNonImplement,
}

type NewResource func(event *v1sync.Event) Resource

var (
	resources = map[string]NewResource{
		Account:      NewAccount,
		Role:         NewRole,
		Microservice: NewMicroservice,
		Instance:     NewInstance,
		Heartbeat:    NewHeartbeat,
		Config:       NewConfig,
		KV:           NewKV,
		Environment:  NewEnvironment,
	}
)

func RegisterResources(name string, nr NewResource) {
	resources[name] = nr
}

func New(event *v1sync.Event) (Resource, *Result) {
	if event == nil {
		return nil, &Result{
			Status:  Fail,
			Message: "event is nil",
		}
	}

	r, ok := resources[event.Subject]
	if !ok {
		return nil, &Result{
			Status:  Skip,
			Message: fmt.Sprintf("resource %s not exist", event.Subject),
		}
	}
	return r(event), nil
}

type operator struct {
	a ActionHandler
}

func newOperator(a ActionHandler) *operator {
	return &operator{
		a: a,
	}
}

func (o *operator) operate(ctx context.Context, action string) *Result {
	var err error
	switch action {
	case sync.CreateAction:
		err = o.a.CreateHandle(ctx)
	case sync.UpdateAction:
		err = o.a.UpdateHandle(ctx)
	case sync.DeleteAction:
		err = o.a.DeleteHandle(ctx)
	default:
		err = fmt.Errorf("action %s is invalid", action)
	}

	r := new(Result)
	r.Status = Success
	if err != nil {
		r.Message = err.Error()
		r.Status = Fail
	}
	return r
}

func NewResult(status int32, message string) *Result {
	return &Result{
		Status:  status,
		Message: message,
	}
}

func FailResult(err error) *Result {
	return NewResult(Fail, err.Error())
}

func SuccessResult() *Result {
	return NewResult(Success, "")
}

func SkipResult() *Result {
	return NewResult(Skip, "")
}

type Result struct {
	EventID string
	Status  int32
	Message string
}

func (r *Result) WithMessage(m string) *Result {
	r.Message = m
	return r
}

func (r *Result) WithEventID(id string) *Result {
	r.EventID = id
	return r
}

func (r *Result) Flag() string {
	return fmt.Sprintf("eventID: %s, status %d:%s, message: %s",
		r.EventID, r.Status, codeDescriber[r.Status], r.Message)
}

func NonImplementResult() *Result {
	return NewResult(NonImplement, "")
}

type FailHandler interface {
	FailHandle(context.Context, int32) (*v1sync.Event, error)
	CanDrop() bool
}

type defaultFailHandler struct {
}

func (d *defaultFailHandler) FailHandle(context.Context, int32) (*v1sync.Event, error) {
	return nil, nil
}

func (d *defaultFailHandler) CanDrop() bool {
	return true
}

type OperateHandler interface {
	LoadCurrentResource(context.Context) *Result
	NeedOperate(context.Context) *Result
	Operate(context.Context) *Result
}

type Resource interface {
	OperateHandler
	FailHandler
}

type ActionHandler interface {
	CreateHandle(context.Context) error
	UpdateHandle(context.Context) error
	DeleteHandle(context.Context) error
}

func newInputParam(input interface{}, callback func()) *inputParam {
	return &inputParam{
		input:    input,
		callback: callback,
	}
}

type inputParam struct {
	input interface{}

	callback func()
}

func newInputLoader(
	event *v1sync.Event,
	create *inputParam,
	update *inputParam,
	delete *inputParam) *inputLoader {
	return &inputLoader{
		event: event,

		createInput:    create.input,
		createCallback: create.callback,

		updateInput:    update.input,
		updateCallback: update.callback,

		deleteInput:    delete.input,
		deleteCallback: delete.callback,
	}
}

type inputLoader struct {
	event *v1sync.Event

	createInput    interface{}
	createCallback func()

	updateInput    interface{}
	updateCallback func()

	deleteInput    interface{}
	deleteCallback func()
}

func (i *inputLoader) loadInputUtil(value interface{}, callback func()) error {
	switch data := value.(type) {
	case *string:
		*data = string(i.event.Value)
	default:
		err := json.Unmarshal(i.event.Value, value)
		if err != nil {
			return err
		}
	}

	if callback != nil {
		callback()
	}
	return nil
}

func (i *inputLoader) loadCreateInput() error {
	return i.loadInputUtil(i.createInput, i.createCallback)
}

func (i *inputLoader) loadUpdateInput() error {
	return i.loadInputUtil(i.updateInput, i.updateCallback)
}

func (i *inputLoader) loadDeleteInput() error {
	return i.loadInputUtil(i.deleteInput, i.deleteCallback)
}

func (i *inputLoader) loadInput() error {
	switch i.event.Action {
	case sync.CreateAction:
		return i.loadCreateInput()
	case sync.UpdateAction:
		return i.loadUpdateInput()
	case sync.DeleteAction:
		return i.loadDeleteInput()
	default:
		return fmt.Errorf("invalid action %s", i.event.Action)
	}
}

type tombstoneLoader interface {
	get(ctx context.Context, req *model.GetTombstoneRequest) (*sync.Tombstone, error)
}

type checker struct {
	curNotNil bool

	event      *v1sync.Event
	updateTime func() (int64, error)
	resourceID string

	tombstoneLoader tombstoneLoader
}

func formatUpdateTimeSecond(src string) (int64, error) {
	updateTime, err := strconv.ParseInt(src, 0, 0)
	if err != nil {
		return 0, err
	}

	return secToNanoSec(updateTime), nil
}

func secToNanoSec(timestamp int64) int64 {
	return timestamp * 1000 * 1000 * 1000
}

func (o *checker) needOperate(ctx context.Context) *Result {
	if o.curNotNil {
		if o.updateTime == nil {
			return nil
		}

		updateTime, err := o.updateTime()
		if err != nil {
			log.Error("get update time failed", err)
			return FailResult(err)
		}
		if updateTime >= o.event.Timestamp {
			return SkipResult()
		}

		return nil
	}

	switch o.event.Action {
	case sync.CreateAction:
		return nil
	case sync.UpdateAction:
		if len(o.resourceID) == 0 {
			return nil
		}
		if o.event.Timestamp+oneDaySecond >= time.Now().Unix() {
			return SkipResult()
		}

		ts, err := o.tombstoneLoader.get(ctx, &model.GetTombstoneRequest{
			ResourceType: o.event.Subject,
			ResourceID:   o.resourceID,
		})
		if err != nil {
			if errors.Is(err, datasource.ErrTombstoneNotExists) {
				return nil
			}

			return FailResult(err)
		}

		if ts.Timestamp > o.event.Timestamp {
			return SkipResult()
		}

		return nil
	case sync.DeleteAction:
		return SkipResult()
	default:
		return FailResult(fmt.Errorf("invalid action %s", o.event.Action))
	}
}

func (o *checker) get(ctx context.Context, req *model.GetTombstoneRequest) (*sync.Tombstone, error) {
	return tombstone.Get(ctx, req)
}

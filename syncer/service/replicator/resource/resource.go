package resource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

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
}

type defaultFailHandler struct {
}

func (d *defaultFailHandler) FailHandle(context.Context, int32) (*v1sync.Event, error) {
	return nil, nil
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
	switch value.(type) {
	case *string:
		v := value.(*string)
		*v = string(i.event.Value)
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
	updateTime func() string
	resourceID string

	tombstoneLoader tombstoneLoader
}

func (o *checker) needOperate(ctx context.Context) *Result {
	if o.curNotNil {
		if o.updateTime == nil {
			return nil
		}

		updateTime, err := strconv.ParseInt(o.updateTime(), 0, 0)
		if err != nil {
			log.Error("parse update time failed", err)
			return FailResult(err)
		}

		updateTime = updateTime * 1000 * 1000 * 1000
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

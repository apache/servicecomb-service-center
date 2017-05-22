//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package cron

import (
	"fmt"
	"github.com/servicecomb/service-center/util"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"time"
)

// Time location, default set by the time.Local (*time.Location)
var loc = time.Local
var attimeRegexp = regexp.MustCompile(`^\d\d:\d\d$`)

// Max number of jobs, hack it if you need.
const MAX_JOBNUM_PER_SCHEDULER = 1000

const (
	JOB_UNIT_TYPE_UNKNOWN int = iota
	JOB_UNIT_TYPE_SECOND
	JOB_UNIT_TYPE_MINUTE
	JOB_UNIT_TYPE_HOUR
	JOB_UNIT_TYPE_DAY
)

func getUnitString(unit int) string {
	switch unit {
	case JOB_UNIT_TYPE_SECOND:
		return "seconds"
	case JOB_UNIT_TYPE_MINUTE:
		return "minutes"
	case JOB_UNIT_TYPE_HOUR:
		return "hours"
	case JOB_UNIT_TYPE_DAY:
		return "days"
	default:
		return "unknown"
	}
}

type Job struct {

	// pause interval * unit bettween runs
	interval uint64

	// pause before first run
	delay uint64

	// the job funcName to run, func[funcName]
	funcName string
	// time units, ,e.g. 'minutes', 'hours'...
	unit int
	// optional time at which this job runs
	atTime string

	// datetime of last run
	lastRunTime time.Time
	// datetime of next run
	nextRunTime time.Time
	// cache the period between last an next run
	period time.Duration

	// Map for the function task store
	funcs map[string]interface{}

	// Map for function and  params of function
	fparams map[string]([]interface{})
}

// Create a new job with the time interval.
func NewJob(intervel uint64) *Job {
	return &Job{
		intervel, 0,
		"",
		JOB_UNIT_TYPE_UNKNOWN,
		"",
		time.Unix(0, 0),
		time.Unix(0, 0), 0,
		make(map[string]interface{}),
		make(map[string]([]interface{})),
	}
}

func (job *Job) String() string {
	return fmt.Sprintf("[JOB]type: %d(%s), interval: %d, delay: %d, at: %s, task: %s", job.unit, getUnitString(job.unit), job.interval, job.delay, job.atTime, job.funcName)
}

// True if the job should be run now
func (job *Job) shouldRun() bool {
	return time.Now().After(job.nextRunTime)
}

//Run the job and immdiately reschedulei it
func (job *Job) run() (result []reflect.Value, err error) {
	if job.delay > 0 {
		job.delay = 0
	}

	f := reflect.ValueOf(job.funcs[job.funcName])
	params := job.fparams[job.funcName]
	if len(params) != f.Type().NumIn() {
		util.LOGGER.Errorf(nil, "The number of param is not adapted(expect: %d, get: %d).", f.Type().NumIn(), len(params))
		return
	}

	util.LOGGER.Debugf("schedule job(%p).", job)
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	job.lastRunTime = time.Now()
	job.scheduleNextRun()
	return
}

// for given function fn , get the name of funciton.
func getFunctionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf((fn)).Pointer()).Name()
}

// Specifies the jobFunc that should be called every time the job runs
//
func (job *Job) Do(jobFun interface{}, params ...interface{}) {
	typ := reflect.TypeOf(jobFun)
	if typ.Kind() != reflect.Func {
		util.LOGGER.Errorf(nil, "only function can be schedule into the job queue.")
		return
	}

	fname := getFunctionName(jobFun)
	job.funcs[fname] = jobFun
	job.fparams[fname] = params
	job.funcName = fname
	//schedule the next run
	util.LOGGER.Warnf(nil, "add job(%p) %s.", job, job)
	job.scheduleNextRun()
}

//	s.Every(1).Days().At("10:30").Do(task)
func (job *Job) At(t string) *Job {
	if !attimeRegexp.MatchString(t) {
		util.LOGGER.Errorf(nil, "invalid input time format '%s'.", t)
		return nil
	}

	hour := int((t[0]-'0')*10 + (t[1] - '0'))
	min := int((t[3]-'0')*10 + (t[4] - '0'))
	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		util.LOGGER.Errorf(nil, "invalid input time '%s'.", t)
		return nil
	}

	job.atTime = fmt.Sprintf("%02d:%02d", hour, min)

	mock := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), int(hour), int(min), 0, 0, loc)

	if job.unit == JOB_UNIT_TYPE_DAY {
		if time.Now().After(mock) {
			job.lastRunTime = mock
		} else {
			job.lastRunTime = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, hour, min, 0, 0, loc)
		}
	}

	return job
}

//Compute the instant when this job should run next
func (job *Job) scheduleNextRun() {
	if job.lastRunTime == time.Unix(0, 0) {
		job.lastRunTime = time.Now()
	}

	if job.delay > 0 {
		// if delay greater than 0, schedule it with delay
		job.nextRunTime = job.lastRunTime.Add(time.Duration(job.delay) * time.Second)
		return
	}

	if job.period != 0 {
		// translate all the units to the Seconds
		job.nextRunTime = job.lastRunTime.Add(job.period * time.Second)
	} else {
		switch job.unit {
		case JOB_UNIT_TYPE_MINUTE:
			job.period = time.Duration(job.interval * 60)
			break
		case JOB_UNIT_TYPE_HOUR:
			job.period = time.Duration(job.interval * 60 * 60)
			break
		case JOB_UNIT_TYPE_DAY:
			job.period = time.Duration(job.interval * 60 * 60 * 24)
			break
		case JOB_UNIT_TYPE_SECOND:
			job.period = time.Duration(job.interval)
		}
		job.nextRunTime = job.lastRunTime.Add(job.period * time.Second)
	}
}

// the follow functions set the job's unit with seconds,minutes,hours...

// Set the unit with seconds
func (j *Job) Seconds() (job *Job) {
	j.unit = JOB_UNIT_TYPE_SECOND
	return j
}

//set the unit with minute
func (j *Job) Minutes() (job *Job) {
	j.unit = JOB_UNIT_TYPE_MINUTE
	return j
}

// Set the unit with hours
func (j *Job) Hours() (job *Job) {
	j.unit = JOB_UNIT_TYPE_HOUR
	return j
}

// Set the job's unit with days
func (j *Job) Days() *Job {
	j.unit = JOB_UNIT_TYPE_DAY
	return j
}

// Class Scheduler, the only data member is the list of jobs.
type Scheduler struct {
	// Array store jobs
	jobs [MAX_JOBNUM_PER_SCHEDULER]*Job

	// Size of jobs which jobs holding.
	size int
}

// Scheduler implements the sort.Interface{} for sorting jobs, by the time nextRunTime

func (s *Scheduler) Len() int {
	return s.size
}

func (s *Scheduler) Swap(i, j int) {
	s.jobs[i], s.jobs[j] = s.jobs[j], s.jobs[i]
}

func (s *Scheduler) Less(i, j int) bool {
	return s.jobs[j].nextRunTime.After(s.jobs[i].nextRunTime)
}

// Create a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{[MAX_JOBNUM_PER_SCHEDULER]*Job{}, 0}
}

// Get the current runnable jobs, which shouldRun is True
func (s *Scheduler) getRunnableJobs() (running_jobs [MAX_JOBNUM_PER_SCHEDULER]*Job, n int) {
	runnableJobs := [MAX_JOBNUM_PER_SCHEDULER]*Job{}
	n = 0
	sort.Sort(s)
	for i := 0; i < s.size; i++ {
		if s.jobs[i].shouldRun() {
			runnableJobs[n] = s.jobs[i]
			n++
		} else {
			break
		}
	}
	return runnableJobs, n
}

// Datetime when the next job should run.
func (s *Scheduler) NextRun() (*Job, time.Time) {
	if s.size <= 0 {
		return nil, time.Now()
	}
	sort.Sort(s)
	return s.jobs[0], s.jobs[0].nextRunTime
}

// Schedule a new periodic job
func (s *Scheduler) Every(interval uint64) *Job {
	job := NewJob(interval)
	s.jobs[s.size] = job
	s.size++
	return job
}

// Schedule a new periodic job
func (s *Scheduler) EveryWithDelay(interval, delay uint64) *Job {
	job := NewJob(interval)
	job.delay = delay
	s.jobs[s.size] = job
	s.size++
	return job
}

// Run all the jobs that are scheduled to run.
func (s *Scheduler) RunPending() {
	runnableJobs, n := s.getRunnableJobs()

	if n != 0 {
		for i := 0; i < n; i++ {
			runnableJobs[i].run()
		}
	}
}

// Remove specific job j
func (s *Scheduler) Remove(j interface{}) {
	i := 0
	for ; i < s.size; i++ {
		if s.jobs[i].funcName == getFunctionName(j) {
			break
		}
	}

	for j := (i + 1); j < s.size; j++ {
		s.jobs[i] = s.jobs[j]
		i++
	}
	s.size = s.size - 1
}

// Delete all scheduled jobs
func (s *Scheduler) Clear() {
	for i := 0; i < s.size; i++ {
		s.jobs[i] = nil
	}
	s.size = 0
}

// Start all the pending jobs
// Add seconds ticker
func (s *Scheduler) Start() chan bool {
	stopped := make(chan bool, 1)
	ticker := time.NewTicker(500 * time.Microsecond)

	go func() {
		for {
			select {
			case <-ticker.C:
				s.RunPending()
			case <-stopped:
				return
			}
		}
	}()

	return stopped
}

// The following methods are shortcuts for not having to
// create a Schduler instance

var defaultScheduler = NewScheduler()

// Schedule a new periodic job
func Every(interval uint64) *Job {
	return defaultScheduler.Every(interval)
}

// Schedule a new periodic job and delay(delay unit: second)
func EveryWithDelay(interval, delay uint64) *Job {
	return defaultScheduler.EveryWithDelay(interval, delay)
}

// Run all jobs that are scheduled to run
func Start() chan bool {
	return defaultScheduler.Start()
}

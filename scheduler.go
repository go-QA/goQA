package goQA

import (
	//	"fmt"
	//"error"
	//"log"
	//"os"
	//"io"
	//"gitorious.org/goqa/goqa.git"
	//"runtime"
	"time"
)

const (
	RUN_STATUS_UNKNOWN = iota
	RUN_STATUS_IDLE
	RUN_STATUS_PAUSED
	RUN_STATUS_RUNNING
	RUN_STATUS_FINISHED
	RUN_STATUS_FAULTED
)

const (
	SCHEDULER_ERROR_UNKNOWN = iota
)
const LAUNCH_DELAY = 1000

//type RunPlan int

//type RunInfoCmd struct {
//	run  *RunPlan
//	comm chan CommandInfo
//}

type SchedulerInfo struct {
	Name       string
	ActiveRuns int
	MaxRuns    int
}

type ScheduleHandler interface {
	SendRunplanInfo(runInfo *RunInfo)
}

type SimpleSchedueHandler struct {
	chRunInfo chan *RunInfo
}

func (sc *SimpleSchedueHandler) Init(chRunInfo chan *RunInfo) {
	sc.chRunInfo = chRunInfo
}
func (sc *SimpleSchedueHandler) SendRunplanInfo(runInfo *RunInfo) {
	sc.chRunInfo <- runInfo
}
type RunplanRouter struct {
	logger       *GoQALog
	m_ScheduleHandler map[string]ScheduleHandler
	chnRuns      chan RunInfo
	chnExit      chan int
}

func (r *RunplanRouter) AddScheduleHandler(name string, schedHandler ScheduleHandler) error {
	r.m_ScheduleHandler[name] = schedHandler
	r.logger.LogDebug("Adding Scedule Handler to Master scheduler '%s'", name)
	return nil
}

func (r *RunplanRouter) StopScheduler(schedID SchedId) error {
	return nil
}

func (r *RunplanRouter) PauseScheduler(schedID SchedId) error {
	return nil
}

func (r *RunplanRouter) GetAvailableSchedulers() []SchedulerInfo {
	return nil
}

func (r *RunplanRouter) Run() {
	go func() {
		alive := true
		for alive {
			select {
			case runInfo := <-r.chnRuns:
				schedHandler := r.getQualifiedScheduleHandler(runInfo)
				schedHandler.SendRunplanInfo(&runInfo)
			}
		}
	}()
}

func (r *RunplanRouter) Init(chnRunplanRouter chan RunInfo, chnExit chan int, logger *GoQALog) {
	r.chnRuns = chnRunplanRouter
	r.chnExit = chnExit
	r.logger = logger
	r.m_ScheduleHandler = make(map[string]ScheduleHandler)
}

func (r *RunplanRouter) getQualifiedScheduleHandler(runInfo RunInfo) ScheduleHandler {
	for name, sched := range r.m_ScheduleHandler {
		r.logger.LogDebug("Scheduler '%s' recieved run ID = %s", name, runInfo.Id)
		return sched
	}
	return nil
}

type SchedId int

type Scheduler interface {
	Start()
	Name() string
	AddRun(run *RunInfo) error
}

type Schedule struct {
	logger       *GoQALog
	exitScheduler bool
	m_runs        map[RunId]*RunPlan
	chnRuns       chan *RunInfo
	Status        int
	name          string
}

func (sc *Schedule) Init(chnRunInfo chan *RunInfo, logger *GoQALog, name string) *Schedule {
	sc.logger = logger
	sc.name = name
	sc.m_runs = make(map[RunId]*RunPlan)
	sc.chnRuns = chnRunInfo
	sc.exitScheduler = false
	sc.Status = RUN_STATUS_IDLE
	return sc
}

func (sc *Schedule) Start() {
	go func() {
		sc.Status = RUN_STATUS_RUNNING
		for sc.exitScheduler == false {
			select {
				case runInfo := <-sc.chnRuns:
					sc.logger.LogDebug("runplan '%s' running....", runInfo.Name)
					if run, err := sc.getQualifiedRun(runInfo); err != nil {
						sc.AddRun(run)
						sc.launchRunplan(run.Id())
					}
				case <-time.After(time.Millisecond * 1000):
			}
			sc.onProcessEvents()
		}
	}()
}

func (sc *Schedule) onProcessEvents() {
	for _, run := range sc.m_runs {
		if run.GetStatus() == RUN_STATUS_FINISHED {
			delete(sc.m_runs, run.Id())
		}
		
	}
	time.Sleep(time.Millisecond * LAUNCH_DELAY)
}

func (sc *Schedule) AddRun(run *RunPlan) error {
	if _, ok := sc.m_runs[run.Id()]; ok {
		sc.m_runs[run.Id()] = run
	}
	return nil	
}

func (sc *Schedule) launchRunplan(id RunId) error {
	run, ok := sc.m_runs[id]
	if !ok {
		return nil
	}
	status := run.GetStatus()
	if status == RUN_STATUS_IDLE || status == RUN_STATUS_PAUSED {
		sc.logger.LogDebug("launchRunplan:: Starting run '%s'", run.Name())
		go run.Start()
		
	} else {
		sc.logger.LogDebug("launchRunplan:: Run Not Launched. Status = %d", run.GetStatus())
	}

	return nil
}

func (sc *Schedule) Name() string {
	return sc.name
}

func (sc *Schedule) getQualifiedRun(runInfo *RunInfo) (*RunPlan, error) {
	runplan := RunPlan{}
	runplan.Init(runInfo, sc.logger)
	return &runplan, nil
}

type RunPlan struct {
	logger *GoQALog
	m_testManager TestManager
	info          RunInfo
	Status        int
}

func (rp *RunPlan) Name() string {
	return rp.info.Name
}

func (rp *RunPlan) Id() RunId {
	return rp.info.Id
}

func (rp *RunPlan) Init(runInfo *RunInfo, logger *GoQALog) {
	rp.logger = logger
	rp.info = *runInfo
	rp.Status = RUN_STATUS_IDLE
}

func (rp *RunPlan) Start() {
	rp.Status = RUN_STATUS_RUNNING
	go func() {
		for loop := 0; loop < 10; loop++ {
			time.Sleep(time.Millisecond * 700)
			rp.logger.LogDebug("Running runplan '%s'....", rp.Id())
		}
		rp.logger.LogDebug("Runplan '%s' Complete", rp.info.Id)
		rp.Status = RUN_STATUS_FINISHED
	}()

}

func (rp *RunPlan) Cancel() {

}

func (rp *RunPlan) Pause() {

}

func (rp *RunPlan) GetStatus() int {
	return rp.Status
}

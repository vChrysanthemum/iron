package iron

import (
	"sync"
	"sync/atomic"
	"time"
)

type TimeTask struct {
	Duration    time.Duration
	Do          func()
	IsRunning   int32
	IsShouldRun int32
}

type TimeTaskMgr struct {
	taskRWMutex sync.RWMutex
	tasks       map[string]*TimeTask
}

func (p *TimeTask) Start() {
	if atomic.LoadInt32(&p.IsRunning) == 1 {
		return
	}

	atomic.StoreInt32(&p.IsRunning, 1)
	for atomic.LoadInt32(&p.IsShouldRun) == 1 {
		p.Do()
		time.Sleep(p.Duration)
	}
	atomic.StoreInt32(&p.IsRunning, 0)
}

func (p *TimeTaskMgr) Init() {
	p.tasks = make(map[string]*TimeTask)
	return
}

func (p *TimeTaskMgr) SetTimeTask(key string, do func(), duration time.Duration) *TimeTask {
	p.taskRWMutex.Lock()
	defer p.taskRWMutex.Unlock()
	var ret = TimeTask{
		Duration:    duration,
		Do:          do,
		IsRunning:   0,
		IsShouldRun: 1,
	}
	var oldtask *TimeTask
	var exists bool

	if oldtask, exists = p.tasks[key]; exists {
		atomic.StoreInt32(&oldtask.IsShouldRun, 0)
		for atomic.LoadInt32(&oldtask.IsRunning) == 0 {
			time.Sleep(time.Second)
		}
	}

	p.tasks[key] = &ret

	return &ret
}

func (p *TimeTaskMgr) startTasks() {
	p.taskRWMutex.RLock()
	defer p.taskRWMutex.RUnlock()
	for _, task := range p.tasks {
		if atomic.LoadInt32(&task.IsRunning) == 0 &&
			atomic.LoadInt32(&task.IsShouldRun) == 1 {
			go task.Start()
		}
	}
}

func (p *TimeTaskMgr) Start() {
	for {
		p.startTasks()
		time.Sleep(time.Second)
	}
}

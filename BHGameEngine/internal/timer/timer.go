package timer

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// TimerTask 定时任务
type TimerTask struct {
	ID       int64         // 任务ID
	Delay    time.Duration // 初始延迟
	Interval time.Duration // 重复间隔（0表示单次执行）
	Callback func()        // 回调函数
	RunAt    time.Time     // 下次执行时间
	Active   bool          // 是否活跃
	mu       sync.RWMutex  // 并发锁
}

// TimerManager 定时器管理器
type TimerManager struct {
	tasks     map[int64]*TimerTask // 任务列表
	priorityQ []*TimerTask         // 优先级队列（按执行时间排序）
	mu        sync.Mutex           // 并发锁
	running   bool                 // 是否运行中
}

func NewTimerManager() *TimerManager {
	return &TimerManager{
		tasks:     make(map[int64]*TimerTask),
		priorityQ: make([]*TimerTask, 0),
	}
}

func (m *TimerManager) AddTask(id int64, delay time.Duration, interval time.Duration, callback func()) {
	task := &TimerTask{
		ID:       id,
		Delay:    delay,
		Interval: interval,
		Callback: callback,
		RunAt:    time.Now().Add(delay),
		Active:   true,
	}

	m.mu.Lock()
	m.tasks[id] = task
	m.insertIntoQueue(task)
	m.mu.Unlock()
}

func (m *TimerManager) RemoveTask(id int64) {
	m.mu.Lock()
	if task, ok := m.tasks[id]; ok {
		task.mu.Lock()
		task.Active = false
		task.mu.Unlock()
		delete(m.tasks, id)
	}
	m.mu.Unlock()
}

func (m *TimerManager) insertIntoQueue(task *TimerTask) {
	pos := 0
	for i, t := range m.priorityQ {
		if task.RunAt.Before(t.RunAt) {
			pos = i
			break
		}
		pos = i + 1
	}
	m.priorityQ = append(m.priorityQ[:pos], append([]*TimerTask{task}, m.priorityQ[pos:]...)...)
}

func (m *TimerManager) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	go func() {
		for {
			m.mu.Lock()
			if !m.running {
				m.mu.Unlock()
				return
			}

			if len(m.priorityQ) == 0 {
				m.mu.Unlock()
				time.Sleep(10 * time.Millisecond)
				continue
			}

			now := time.Now()
			task := m.priorityQ[0]

			if task.RunAt.After(now) {
				sleepTime := task.RunAt.Sub(now)
				m.mu.Unlock()
				time.Sleep(sleepTime)
				continue
			}

			m.priorityQ = m.priorityQ[1:]
			m.mu.Unlock()

			task.mu.RLock()
			active := task.Active
			task.mu.RUnlock()

			if !active {
				continue
			}

			go func() {
				defer func() {
					if r := recover(); r != nil {
						logrus.Error("Timer task panic: ", r)
					}
				}()
				task.Callback()
			}()

			if task.Interval > 0 {
				m.mu.Lock()
				task.RunAt = time.Now().Add(task.Interval)
				m.insertIntoQueue(task)
				m.mu.Unlock()
			} else {
				m.mu.Lock()
				delete(m.tasks, task.ID)
				m.mu.Unlock()
			}
		}
	}()
}

func (m *TimerManager) Stop() {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()
}

func (m *TimerManager) GetTaskCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.tasks)
}

func (m *TimerManager) UpdateTaskDelay(id int64, delay time.Duration) {
	m.mu.Lock()
	task, ok := m.tasks[id]
	if ok {
		task.mu.Lock()
		task.Delay = delay
		task.RunAt = time.Now().Add(delay)
		task.mu.Unlock()

		for i, t := range m.priorityQ {
			if t.ID == id {
				m.priorityQ = append(m.priorityQ[:i], m.priorityQ[i+1:]...)
				break
			}
		}
		m.insertIntoQueue(task)
	}
	m.mu.Unlock()
}

// RegionTimerManager 区域定时器管理器
type RegionTimerManager struct {
	timers map[string]*TimerManager // 区域定时器列表（区域ID->定时器）
	mu     sync.RWMutex             // 并发锁
}

func NewRegionTimerManager() *RegionTimerManager {
	return &RegionTimerManager{
		timers: make(map[string]*TimerManager),
	}
}

func (r *RegionTimerManager) GetRegionTimer(regionID string) *TimerManager {
	r.mu.RLock()
	timer, ok := r.timers[regionID]
	r.mu.RUnlock()

	if !ok {
		r.mu.Lock()
		timer = NewTimerManager()
		r.timers[regionID] = timer
		timer.Start()
		r.mu.Unlock()
	}

	return timer
}

func (r *RegionTimerManager) RemoveRegionTimer(regionID string) {
	r.mu.Lock()
	if timer, ok := r.timers[regionID]; ok {
		timer.Stop()
		delete(r.timers, regionID)
	}
	r.mu.Unlock()
}

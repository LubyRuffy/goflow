package web

import (
	"github.com/LubyRuffy/goflow"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	globalTaskMonitor = newTaskMonitor()
)

type msgItem struct {
	ts  string
	msg string
}

type taskInfo struct {
	monitor         *taskMonitor
	runner          *goflow.PipeRunner
	taskId          string
	astCode         string // 运行的代码
	msgs            []msgItem
	started         time.Time
	ended           time.Time
	html            string
	actionIDRunning string // 当前运行的actionID
	finished        bool
	sync.Mutex
}

func (t *taskInfo) finish() {
	t.ended = time.Time{}
	t.finished = true

	go func() {
		select {
		case <-time.After(1 * time.Minute):
			// 1分钟后删除
			t.monitor.del(t.taskId)
		}
	}()
}

func (t *taskInfo) addMsg(msg string) {
	if t.finished {
		return
	}
	t.Lock()
	defer t.Unlock()
	t.msgs = append(t.msgs, msgItem{
		ts:  strconv.FormatInt(time.Now().Unix(), 10),
		msg: msg,
	})
}

// 返回列表和最后时间戳
func (t *taskInfo) receiveMsgs(ts string) ([]string, string) {
	var msgs []string
	found := false
	if len(ts) == 0 {
		found = true
	}
	lastTimeStamp := ""
	for i := range t.msgs {
		if t.msgs[i].ts != ts {
			if found {
				msgs = append(msgs, t.msgs[i].msg)
				lastTimeStamp = t.msgs[i].ts
			} else {
				continue
			}
		} else {
			found = true
		}
	}
	return msgs, lastTimeStamp
}

type taskMonitor struct {
	m sync.Map
}

func newTaskMonitor() *taskMonitor {
	return &taskMonitor{}
}

func (t *taskMonitor) del(taskId string) {
	t.m.Delete(taskId)
}

func (t *taskMonitor) new(code string) *taskInfo {
	tid := uuid.New().String()
	ti := &taskInfo{
		taskId:  tid,
		astCode: code,
		started: time.Now(),
		monitor: t,
	}
	t.m.Store(tid, ti)
	return ti
}

func (t *taskMonitor) addMsg(taskid string, msg string) {
	if task, ok := t.m.Load(taskid); ok {
		task.(*taskInfo).addMsg(msg)
	}
}

func (t *taskMonitor) receiveMsgs(taskid string, ts string) ([]string, string) {
	if task, ok := t.m.Load(taskid); ok {
		return task.(*taskInfo).receiveMsgs(ts)
	}
	return nil, ""
}

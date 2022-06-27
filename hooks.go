package goflow

import (
	"github.com/sirupsen/logrus"
	"log"
)

// Hooks 消息通知
type Hooks struct {
	OnWorkflowFinished func(pt *PipeTask)                                           // 一个workflow完成时的处理
	OnWorkflowStart    func(funcName string, actionID string)                       // 一个workflow完成时的处理
	OnLog              func(level logrus.Level, format string, args ...interface{}) // 日志通知
	OnGetObject        func(name string) (interface{}, bool)                        // 底层要获取上层定义的对象
	OnProgress         func(p float64)                                              // 进度提示
}

var (
	defaultHooks = &Hooks{
		OnWorkflowFinished: func(pt *PipeTask) {

		},
		OnWorkflowStart: func(funcName string, actionID string) {

		},
		OnLog: func(level logrus.Level, format string, args ...interface{}) {

		},
		OnGetObject: func(name string) (interface{}, bool) {
			return nil, false
		},
		OnProgress: func(p float64) {
			log.Printf("progress: %f%%\n", p)
		},
	}
)

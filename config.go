package goflow

import (
	"github.com/sirupsen/logrus"
)

// WithHooks user defined hooks
func (p *PipeRunner) WithHooks(hooks *Hooks) *PipeRunner {
	p.hooks = hooks
	if p.hooks.OnWorkflowStart == nil {
		p.hooks.OnWorkflowStart = func(funcName string, actionID string) {
		}
	}
	if p.hooks.OnWorkflowFinished == nil {
		p.hooks.OnWorkflowFinished = func(pt *PipeTask) {
		}
	}
	if p.hooks.OnLog == nil {
		p.hooks.OnLog = func(level logrus.Level, format string, args ...interface{}) {
		}
	}
	if p.hooks.OnProgress == nil {
		p.hooks.OnProgress = func(p float64) {
		}
	}
	if p.hooks.OnGetObject == nil {
		p.hooks.OnGetObject = func(name string) (interface{}, bool) {
			return nil, false
		}
	}

	return p
}

// WithParent from hook
func (p *PipeRunner) WithParent(parent *PipeRunner) *PipeRunner {
	p.Parent = parent
	return p
}

// WithUserFunction Function to register
func (p *PipeRunner) WithUserFunction(funcs ...[]interface{}) *PipeRunner {
	p.registerFunctions(funcs...)
	return p
}

// WithObject register object
func (p *PipeRunner) WithObject(name string, obj interface{}) *PipeRunner {
	p.objects.Store(name, obj)
	return p
}

// WithDebug open debug
func (p *PipeRunner) WithDebug(level logrus.Level) *PipeRunner {
	p.logger.Level = level
	return p
}

// WithWebHook webhook setting
func (p *PipeRunner) WithWebHook(hookURL string) *PipeRunner {
	p.WebHook = hookURL
	return p
}

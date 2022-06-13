package goflow

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/sirupsen/logrus"
)

// WithHooks user defined hooks
func (p *PipeRunner) WithHooks(hooks *Hooks) *PipeRunner {
	p.hooks = hooks
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

// WithAST Function to register
func (p *PipeRunner) WithAST(ast *workflowast.Parser) *PipeRunner {
	p.ast = ast
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

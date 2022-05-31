package goflow

import "github.com/LubyRuffy/goflow/workflowast"

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

// WithObject Function to register
func (p *PipeRunner) WithObject(name string, obj interface{}) *PipeRunner {
	p.objects.Store(name, obj)
	return p
}

package agent

import (
	"context"

	"spawn.dev/pkg/capability"
	"spawn.dev/pkg/llm"
)

// ExecutionContext holds runtime state and memory handles.
type ExecutionContext struct {
	WorkDir string
	Env     map[string]string
	Secrets map[string]string

	Messages []llm.Message

	VectorStore capability.VectorStore
	GraphStore  capability.GraphStore
	KVStore     capability.KVStore

	ToolCache map[string]interface{}

	ctx    context.Context
	cancel context.CancelFunc
}

// NewExecutionContext creates an execution context with cancellation.
func NewExecutionContext(parent context.Context, workDir string) *ExecutionContext {
	ctx, cancel := context.WithCancel(parent)
	return &ExecutionContext{
		WorkDir:   workDir,
		Env:       map[string]string{},
		Secrets:   map[string]string{},
		ToolCache: map[string]interface{}{},
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Done returns cancellation signal.
func (e *ExecutionContext) Done() <-chan struct{} {
	if e == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return e.ctx.Done()
}

// Cancel cancels the execution context.
func (e *ExecutionContext) Cancel() {
	if e != nil && e.cancel != nil {
		e.cancel()
	}
}

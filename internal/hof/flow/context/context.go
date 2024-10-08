/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

// This package provides a context for Tasks
// and a registry for their usage in flows.
package context

import (
	gocontext "context"
	"io"
	"sync"

	"cuelang.org/go/cue"

	"github.com/opentofu/opentofu/internal/hof/flow/task"
)

// A Context provides context for running a task.
type Context struct {
	RootValue cue.Value
	GoContext gocontext.Context

	FlowStack []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// the value worked on
	Value cue.Value
	Error error

	// debug / internal
	Verbosity int
	Plan      bool
	Gist      bool
	Apply     bool
	Init      bool
	Destroy   bool

	Middlewares  []Middleware
	TaskRegistry *sync.Map

	// BOOKKEEPING
	Tasks *sync.Map

	// experimental
	BaseTask *task.BaseTask

	// Middleware
	Pools *sync.Map

	// how can the below become middleware, extensions, plugin?

	// Global (for this context, tbd shared) lock around CUE evaluator
	CUELock *sync.Mutex

	// map of cue.Values
	ValStore *sync.Map

	// map of chan?
	Mailbox *sync.Map

	// channels for
	// - stats & progress
	// CUE context
	CueContext *cue.Context

	// output vars
	GlobalVars *sync.Map
	// store flow errors and warnings
	FlowErrors   []string
	FlowWarnings []string
}

func New() *Context {
	return &Context{
		GoContext:    gocontext.Background(),
		FlowStack:    []string{},
		CUELock:      new(sync.Mutex),
		ValStore:     new(sync.Map),
		Mailbox:      new(sync.Map),
		Middlewares:  make([]Middleware, 0),
		TaskRegistry: new(sync.Map),
		Tasks:        new(sync.Map),
		Pools:        new(sync.Map),
		CueContext:   nil,
		GlobalVars:   new(sync.Map),
		FlowErrors:   []string{},
		FlowWarnings: []string{},
	}
}

func Copy(ctx *Context) *Context {
	return &Context{
		RootValue: ctx.RootValue,
		GoContext: ctx.GoContext,
		FlowStack: ctx.FlowStack,

		Stdin:  ctx.Stdin,
		Stdout: ctx.Stdout,
		Stderr: ctx.Stderr,

		Verbosity: ctx.Verbosity,

		CUELock:  ctx.CUELock,
		Mailbox:  ctx.Mailbox,
		ValStore: ctx.ValStore,

		Middlewares:  ctx.Middlewares,
		TaskRegistry: ctx.TaskRegistry,
		Tasks:        ctx.Tasks,
		Pools:        ctx.Pools,
		Apply:        ctx.Apply,
		Plan:         ctx.Plan,
		Gist:         ctx.Gist,
		Init:         ctx.Init,
		Destroy:      ctx.Destroy,
		CueContext:   ctx.CueContext,
		GlobalVars:   ctx.GlobalVars,
	}
}

func (C *Context) Use(m Middleware) {
	C.Middlewares = append(C.Middlewares, m)
}

// Register registers a task for cue commands.
func (C *Context) Register(key string, f RunnerFunc) {
	C.TaskRegistry.Store(key, f)
}

// Lookup returns the RunnerFunc for a key.
func (C *Context) Lookup(key string) RunnerFunc {
	v, ok := C.TaskRegistry.Load(key)
	if !ok {
		return nil
	}
	return v.(RunnerFunc)
}

// AddError adds an error to the flow errors.
func (C *Context) AddError(errString string) {
	C.FlowErrors = append(C.FlowErrors, errString)
}

// AddWarning adds a warning to the flow warnings.
func (C *Context) AddWarning(warning string) {
	C.FlowWarnings = append(C.FlowWarnings, warning)
}

// Middleware to apply to RunnerFuncs
// should wrap and call Run of the passed RunnerFunc?
type Middleware interface {
	Apply(*Context, RunnerFunc) RunnerFunc
}

// A RunnerFunc creates a Runner.
type RunnerFunc func(v cue.Value) (Runner, error)

// A Runner defines a task type.
type Runner interface {
	// Runner runs given the current value and returns a new value which is to
	// be unified with the original result.
	Run(ctx *Context) (results interface{}, err error)
}

// TFContext is a type that includes ParsedVariables which is a map of string keys to another map of string keys and cty.Value values.
type TFContext struct {
	ParsedVariables *sync.Map
}

// NewTFContext is a constructor for TFContext.
func NewTFContext(parsedVars *sync.Map) *TFContext {
	return &TFContext{
		ParsedVariables: parsedVars,
	}
}

/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package tasker

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/token"
	cueflow "cuelang.org/go/tools/flow"

	flowctx "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/flow/task"
	"github.com/opentofu/opentofu/internal/hof/lib/cuetils"
	"github.com/opentofu/opentofu/internal/hof/lib/hof"
)

var debug = false

func NewTasker(ctx *flowctx.Context) cueflow.TaskFunc {
	// This function implements the Runner interface.
	// It parses Cue values, you will see all of them recursively

	return func(val cue.Value) (cueflow.Runner, error) {
		// fmt.Println("Tasker:", val.Path())

		// what's going on here?
		// this the root value? (so def a flow, so not a task)
		if len(val.Path().Selectors()) == 0 {
			return nil, nil
		}

		node, err := hof.ParseHof[any](val)
		if err != nil {
			return nil, err
		}
		if node == nil {
			return nil, nil
		}
		if node.Hof.Flow.Task == "" {
			return nil, nil
		}
		//if node.Hof.Flow.Task == "nest" {
		//  fmt.Println("New Tasker NEST", node.Hof.Path, node.Hof.Label)
		//}

		return makeTask(ctx, node)
	}
}

func makeTask(ctx *flowctx.Context, node *hof.Node[any]) (cueflow.Runner, error) {

	taskId := node.Hof.Flow.Task

	//taskName := node.Hof.Flow.Name
	//fmt.Println("makeTask:", taskId, taskName, node.Hof.Path, node.Hof.Flow.Root, node.Parent)

	// lookup context.RunnerFunc
	runnerFunc := ctx.Lookup(taskId)
	if runnerFunc == nil {
		return nil, fmt.Errorf("unknown task: %q at %q", taskId, node.Value.Path())
	}

	// Note, we apply this in the reverse order so that the Use order is like a stack
	// (i.e. the first is the most outer, which is typical for how these work for servers
	// apply plugin / middleware
	for i := len(ctx.Middlewares) - 1; i >= 0; i-- {
		ware := ctx.Middlewares[i]
		runnerFunc = ware.Apply(ctx, runnerFunc)
	}

	// some way to validate task against it's schema
	// (1) schemas self register
	// (2) here, we lookup schemas by taskId
	// (3) use custom Require (or other validator)

	// create hof task from val
	// these live under /flow/tasks
	// and are of type context.RunnerFunc
	T, err := runnerFunc(node.Value)
	if err != nil {
		return nil, err
	}

	// do per-task setup / common base / initial value / bookkeeping
	bt := task.NewBaseTask(node)
	ctx.Tasks.Store(bt.ID, bt)

	// wrap our RunnerFunc with cue/flow RunnerFunc
	return cueflow.RunnerFunc(func(t *cueflow.Task) error {
		fmt.Println("makeTask.func()", t.Index(), t.Path())

		// why do we need a copy?
		// maybe for local Value / CurrTask
		c := flowctx.Copy(ctx)

		c.Value = t.Value()
		node, err := hof.ParseHof[any](c.Value)
		if err != nil {
			return err
		}

		// Inject variables before running the task
		injectedNode, err := injectVariables(node.Value, c.GlobalVars)
		if err != nil {
			return fmt.Errorf("error injecting variables: %v", err)
		}
		c.Value = c.CueContext.BuildExpr(injectedNode)

		fmt.Println("Injected value: %v", c.Value)

		if node.Hof.Flow.Print.Level > 0 && node.Hof.Flow.Print.Before {
			pv := c.Value.LookupPath(cue.ParsePath(node.Hof.Flow.Print.Path))
			if node.Hof.Path == "" {
				fmt.Printf("%s", node.Hof.Flow.Print.Path)
			} else if node.Hof.Flow.Print.Path == "" {
				fmt.Printf("%s", node.Hof.Path)
			} else {
				fmt.Printf("%s.%s", node.Hof.Path, node.Hof.Flow.Print.Path)
			}
			fmt.Printf(": %#v\n", pv)
		}

		c.BaseTask = bt

		// fmt.Println("MAKETASK", taskId, c.FlowStack, c.Value.Path())
		// fmt.Printf("%# v\n", c.Value)

		bt.CueTask = t
		bt.Start = c.Value
		// TODO, we should remove this next line, and only set Final at the end
		bt.Final = c.Value

		// run the hof task
		bt.AddTimeEvent("run.beg")
		// (update)
		value, rerr := T.Run(c)
		bt.AddTimeEvent("run.end")
		updateGlobalVars(c, bt.Final)

		if value != nil {
			// fmt.Println("FILL:", taskId, c.Value.Path(), t.Value(), value)
			bt.AddTimeEvent("fill.beg")

			//if node.Hof.Flow.Task == "nest" || node.Hof.Flow.Task == "api.Call" {
			//  fmt.Println("FILL:", taskId, c.Value.Path(), value)
			//}
			err = t.Fill(value)
			bt.Final = t.Value()
			bt.AddTimeEvent("fill.end")

			// fmt.Println("FILL:", taskId, c.Value.Path(), t.Value(), value)
			if err != nil {
				c.Error = err
				bt.Error = err
				return err
			}

			//if node.Hof.Flow.Print.Level > 0 && !node.Hof.Flow.Print.Before {
			//  // pv := bt.Final.LookupPath(cue.ParsePath(node.Hof.Flow.Print.Path))
			//  fmt.Printf("%s.%s: %# v\n", node.Hof.Path, node.Hof.Flow.Print.Path, value)
			//}
			// --------------------------------
		}

		if rerr != nil {
			rerr = fmt.Errorf("in %q\n%v\n%+v", c.Value.Path(), cuetils.ExpandCueError(rerr), value)
			// fmt.Println("RunnerRunc Error:", err)
			c.Error = rerr
			bt.Error = rerr
			return rerr
		}

		return nil
	}), nil
}
func injectVariables(value cue.Value, globalVars map[string]interface{}) (ast.Expr, error) {
	f := value.Syntax(cue.Final()).(ast.Expr)

	imports, _ := value.LookupPath(cue.ParsePath("imports")).Fields()
	importMap := make(map[string]string)
	for imports.Next() {
		importMap[imports.Label()], _ = imports.Value().String()
	}

	injectedNode := astutil.Apply(f, nil, func(c astutil.Cursor) bool {
		n := c.Node()

		switch x := n.(type) {
		case *ast.Field:
			for _, attr := range x.Attrs {
				if strings.HasPrefix(attr.Text, "@runinject") {
					varName := parseRunInjectAttr(attr.Text)
					if importPath, ok := importMap[varName]; ok {
						if val, ok := lookupNestedValue(globalVars, importPath); ok {
							x.Value = &ast.BasicLit{
								Kind:  token.STRING,
								Value: fmt.Sprintf("%q", val),
							}
						}
					}
				}
			}
		}
		return true
	})

	return injectedNode.(ast.Expr), nil
}

func parseRunInjectAttr(attrText string) string {
	attrText = strings.TrimPrefix(attrText, "@runinject(")
	attrText = strings.TrimSuffix(attrText, ")")
	return strings.Trim(attrText, "\"")
}

func lookupNestedValue(m map[string]interface{}, key string) (string, bool) {
	parts := strings.Split(key, ".")
	var current interface{} = m

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, ok := v[part]; ok {
				current = val
			} else {
				return "", false
			}
		default:
			return "", false
		}
	}

	return fmt.Sprint(current), true
}

func updateGlobalVars(ctx *flowctx.Context, taskOutput cue.Value) {
	taskPath := taskOutput.Path().String()
	outputs, _ := taskOutput.LookupPath(cue.ParsePath("outputs")).List()

	for i := 0; outputs.Next(); i++ {
		outputVar, _ := outputs.Value().String()
		if value := taskOutput.LookupPath(cue.ParsePath(outputVar)); value.Exists() {
			setNestedValue(ctx.GlobalVars, fmt.Sprintf("%s.outputs.%s", taskPath, outputVar), fmt.Sprint(value))
		}
	}
}

func setNestedValue(m map[string]interface{}, key string, value string) {
	parts := strings.Split(key, ".")
	current := m

	for _, part := range parts[:len(parts)-1] {
		if _, ok := current[part]; !ok {
			current[part] = make(map[string]interface{})
		}
		if nested, ok := current[part].(map[string]interface{}); ok {
			current = nested
		} else {
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		}
	}

	current[parts[len(parts)-1]] = value
}

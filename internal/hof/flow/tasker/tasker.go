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

		// fmt.Println("Injected value: %v", c.Value)

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
			if cueValue, ok := value.(cue.Value); ok {
				updateGlobalVars(c, bt, cueValue)
			} else {
				return fmt.Errorf("expected cue.Value, got %T", value)
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

	inputs, _ := value.LookupPath(cue.ParsePath("inputs")).Fields()
	inputMap := make(map[string]string)
	for inputs.Next() {
		inputName := inputs.Label()
		inputPath, _ := inputs.Value().String() // @todo: handle error
		if val, ok := lookupNestedValue(globalVars, inputPath); ok {
			inputMap[inputName] = val
		} else {
			return nil, fmt.Errorf("failed to resolve inputs %s: %s", inputName, inputPath)
		}
	}

	injectedNode := astutil.Apply(f, nil, func(c astutil.Cursor) bool {
		n := c.Node()

		switch x := n.(type) {
		case *ast.Field:
			for _, attr := range x.Attrs {
				if strings.HasPrefix(attr.Text, "@runinject") {
					varName := parseRunInjectAttr(attr.Text)
					if val, ok := inputMap[varName]; ok {
						x.Value = &ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf("%q", val),
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
	current := m

	for _, part := range parts {
		if val, ok := current[part]; ok {
			if nextMap, isMap := val.(map[string]interface{}); isMap {
				current = nextMap
			} else {
				return fmt.Sprint(val), true
			}
		} else {
			return "", false
		}
	}
	return "", false
}

func updateGlobalVars(ctx *flowctx.Context, bt *task.BaseTask, value cue.Value) {
	taskPath := bt.ID // Use the task ID as the path

	outputsValue := bt.Final.LookupPath(cue.ParsePath("outputs"))
	outValue := value.LookupPath(cue.ParsePath("out"))

	// debugPrintCueValue("outValue", outValue)

	if outputsValue.Exists() {
		switch outputsValue.Kind() {
		case cue.StructKind:
			iter, _ := outputsValue.Fields()
			for iter.Next() {
				outputVar := iter.Label()
				outputPath, _ := iter.Value().String()
				fmt.Printf("Processing output: %s with path: %s\n", outputVar, outputPath)
				processOutput(ctx, taskPath, outputVar, outputPath, outValue)
			}
		default:
			fmt.Printf("Unexpected outputs kind: %v\n", outputsValue.Kind())
		}
	} else {
		fmt.Println("No outputs defined for this task")
	}
}

func processOutput(ctx *flowctx.Context, taskPath, outputVar, outputPath string, outValue cue.Value) {
	if actualValue := fetchActualValue(outValue, outputPath); actualValue.Exists() {
		// Debug: Print the found output value
		// debugPrintCueValue(fmt.Sprintf("Found value for %s", outputVar), actualValue)
		fullPath := fmt.Sprintf("tasks.%s.outputs.%s", taskPath, outputVar)
		setNestedValue(ctx.GlobalVars, fullPath, formatValue(actualValue))
	} else {
		// Debug: Print when output value is not found
		fmt.Printf("Value not found for output: %s at path: %s\n", outputVar, outputPath)

		// Additional debugging: print the structure of outValue
		// debugPrintCueValue("outValue structure", outValue)
	}
}

func setNestedValue(m map[string]interface{}, key string, value interface{}) {
	parts := strings.Split(key, ".")
	current := m

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
		} else {
			if _, ok := current[part]; !ok {
				current[part] = make(map[string]interface{})
			}
			current = current[part].(map[string]interface{})
		}
	}
}

func fetchActualValue(value cue.Value, path string) cue.Value {
	parts := strings.Split(path, ".")

	// this is a hack to handle cases where the key is a nested key
	if len(parts) < 2 {
		// If there's only one part, just do a direct lookup
		return value.LookupPath(cue.ParsePath(path))
	}

	// Combine all parts except the last one as the main key
	mainKey := strings.Join(parts[:len(parts)-1], ".")
	lastKey := parts[len(parts)-1]

	// First, try to lookup the main key as a whole
	mainValue := value.LookupPath(cue.ParsePath(mainKey))
	if !mainValue.Exists() {
		// If that fails, try to lookup the main key as a quoted string
		mainValue = value.LookupPath(cue.ParsePath(fmt.Sprintf("%q", mainKey)))
	}

	if !mainValue.Exists() {
		fmt.Printf("Main key not found: %s\n", mainKey)
		// debugPrintCueValue("Current value structure", value)
		return cue.Value{}
	}

	fmt.Printf("Looking up last key: %s\n", lastKey)
	result := mainValue.LookupPath(cue.ParsePath(lastKey))
	if !result.Exists() {
		fmt.Printf("Last key not found: %s\n", lastKey)
		// debugPrintCueValue("Main value structure", mainValue)
		return cue.Value{}
	}

	return result
}

func formatValue(v cue.Value) interface{} {
	switch v.Kind() {
	case cue.StringKind:
		str, _ := v.String()
		return str
	case cue.IntKind:
		i, _ := v.Int64()
		return i
	case cue.FloatKind:
		f, _ := v.Float64()
		return f
	case cue.BoolKind:
		b, _ := v.Bool()
		return b
	case cue.StructKind:
		m := make(map[string]interface{})
		iter, _ := v.Fields()
		for iter.Next() {
			m[iter.Label()] = formatValue(iter.Value())
		}
		return m
	case cue.ListKind:
		var list []interface{}
		iter, _ := v.List()
		for iter.Next() {
			list = append(list, formatValue(iter.Value()))
		}
		return list
	default:
		return fmt.Sprint(v)
	}
}

// func debugPrintCueValue(label string, v cue.Value) {
// 	fmt.Printf("--- Debug: %s ---\n", label)
// 	fmt.Printf("Kind: %v\n", v.Kind())

// 	switch v.Kind() {
// 	case cue.StructKind:
// 		fmt.Println("Structure:")
// 		iter, _ := v.Fields()
// 		for iter.Next() {
// 			fmt.Printf("  %s: %v\n", iter.Label(), iter.Value())
// 		}
// 	case cue.ListKind:
// 		fmt.Println("List:")
// 		list, _ := v.List()
// 		for list.Next() {
// 			fmt.Printf("  %v\n", list.Value())
// 		}
// 	default:
// 		fmt.Printf("Value: %v\n", v)
// 	}

// 	// Print CUE syntax representation
// 	syn := v.Syntax(cue.Final())
// 	bytes, _ := cueformat.Node(syn)
// 	fmt.Printf("CUE syntax:\n%s\n", string(bytes))

// 	fmt.Println("------------------------")
// }

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
	"strconv"
	"strings"

	"github.com/itchyny/gojq"

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
		//fmt.Println("makeTask.func()", t.Index(), t.Path())

		// why do we need a copy?
		// maybe for local Value / CurrTask
		c := flowctx.Copy(ctx)

		c.Value = t.Value()
		node, err := hof.ParseHof[any](c.Value)
		if err != nil {
			return err
		}

		// Inject variables before running the task
		// (only if we are applying)
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
				updateGlobalVars(c, cueValue)
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
	// Process @preinject attributes before @runinject
	injectedNode := astutil.Apply(f, nil, func(c astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.Field:
			for _, attr := range x.Attrs {
				if strings.HasPrefix(attr.Text, "@runinject") {
					varName := parseRunInjectAttr(attr.Text)
					if val, ok := globalVars[varName]; ok {
						x.Value = createASTNodeForValue(val)
					} else {
						fmt.Printf("Warning: Unable to find runtime alias: %v\n", varName)
						// print global vars
						fmt.Printf("Please see if it's mispelt.\n")
						// Print all globalVars as key-value pairs
						fmt.Println("Currently available runtime aliases: ")
						for k, v := range globalVars {
							fmt.Printf("%s: %v\n", k, v)
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

func createASTNodeForValue(val interface{}) ast.Expr {
	switch v := val.(type) {
	case string:
		return &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(v)}
	case int:
		return &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(v)}
	case float64:
		return &ast.BasicLit{Kind: token.FLOAT, Value: strconv.FormatFloat(v, 'f', -1, 64)}
	case bool:
		if v {
			return &ast.BasicLit{Kind: token.TRUE, Value: "true"}
		} else {
			return &ast.BasicLit{Kind: token.FALSE, Value: "false"}
		}
	case []interface{}:
		elts := make([]ast.Expr, len(v))
		for i, item := range v {
			elts[i] = createASTNodeForValue(item)
		}
		return &ast.ListLit{Elts: elts}
	case map[string]interface{}:
		fields := make([]ast.Decl, 0, len(v))
		for key, value := range v {
			fields = append(fields, &ast.Field{
				Label: ast.NewString(key),
				Value: createASTNodeForValue(value),
			})
		}
		return &ast.StructLit{Elts: fields}
	default:
		// For any other types, convert to string as a fallback
		return &ast.BasicLit{Kind: token.NULL, Value: ast.NewNull().Value}
	}
}

func updateGlobalVars(ctx *flowctx.Context, value cue.Value) {
	outputsValue := value.LookupPath(cue.ParsePath("outputs"))
	outValue := value.LookupPath(cue.ParsePath("out"))
	// Check if outputsValue is null
	if !outputsValue.Exists() {
		return
	}

	// Parse outValue into a Go object
	var outData interface{}
	if err := outValue.Decode(&outData); err != nil {
		// Instead of returning immediately, let's try to work with the partial data
		outData = convertCueToInterface(outValue)
	}

	if outputsValue.Exists() {
		switch outputsValue.Kind() {
		case cue.ListKind:
			iter, _ := outputsValue.List()
			for iter.Next() {
				outputDef := iter.Value()
				alias, _ := outputDef.LookupPath(cue.ParsePath("alias")).String()
				jqPath, _ := outputDef.LookupPath(cue.ParsePath("path")).String()
				actualValue := processOutput(alias, jqPath, outData)
				ctx.GlobalVars[alias] = actualValue
			}
		default:
			fmt.Printf("Unexpected outputs kind: %v\n", outputsValue.Kind())
		}
	} else {
		fmt.Println("No outputs defined for this task")
	}
}

func processOutput(alias, jqPath string, outData interface{}) interface{} {
	actualValue, ok := queryJQ(outData, jqPath)
	if !ok {
		fmt.Printf("Value not found at path: %s for alias: %s\n", jqPath, alias)
		return nil
	}

	// Determine if the query should return an array
	if shouldReturnArray(jqPath) {
		// Ensure the result is always an array
		if slice, isSlice := actualValue.([]interface{}); isSlice {
			return slice
		}
		// If it's not already a slice, wrap it in one
		return []interface{}{actualValue}
	}

	// For non-array queries, return the value as-is
	return actualValue
}

func queryJQ(data interface{}, jqPath string) (interface{}, bool) {
	query, err := gojq.Parse(jqPath)
	if err != nil {
		fmt.Printf("Error parsing JQ query (%s): %v\n", jqPath, err)
		return nil, false
	}

	iter := query.Run(data)
	var results []interface{}

	for {
		result, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := result.(error); isErr {
			fmt.Printf("Error during JQ query execution: %v\n", err)
			continue
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, false
	}
	if len(results) == 1 {
		return results[0], true
	}
	return results, true
}

func shouldReturnArray(jqPath string) bool {
	// Check if the query ends with '[]' or contains '[*]'
	return strings.HasSuffix(jqPath, "[]") || strings.Contains(jqPath, "[*]") ||
		// Check for array operations like '.[]'
		strings.Contains(jqPath, ".[]") ||
		// Check for patterns like ".module.vpc.aws_subnet.public[].id"
		strings.Contains(jqPath, "[].") ||
		// Check for map operations that result in arrays
		strings.Contains(jqPath, "| keys") || strings.Contains(jqPath, "| values") ||
		// Add more conditions as needed for other array-producing operations
		strings.Contains(jqPath, "| to_entries")
}

// Helper function to convert CUE value to interface{} with support for nested maps to arrays
func convertCueToInterface(v cue.Value) interface{} {
	switch v.Kind() {
	case cue.StructKind:
		result := make(map[string]interface{})
		iter, _ := v.Fields()
		for iter.Next() {
			result[iter.Label()] = convertCueToInterface(iter.Value())
		}
		// If the struct has numeric keys, convert to an array
		if isNumericKeys(result) {
			return mapToArray(result)
		}
		return result
	case cue.ListKind:
		var result []interface{}
		iter, _ := v.List()
		for iter.Next() {
			result = append(result, convertCueToInterface(iter.Value()))
		}
		return result
	default:
		if !v.IsConcrete() {
			// Log more information about the non-concrete value
			return fmt.Sprintf("_non_concrete(%s)", v.Path())
		}
		var result interface{}
		if err := v.Decode(&result); err != nil {
			fmt.Printf("Error decoding value at path %v: %v\n", v.Path(), err)
			return v
		}
		return result
	}
}

// Check if a map has all numeric keys
func isNumericKeys(data map[string]interface{}) bool {
	for k := range data {
		if _, err := strconv.Atoi(k); err != nil {
			return false
		}
	}
	return true
}

// Convert a map with numeric keys into an array, handling nested maps recursively
func mapToArray(data map[string]interface{}) []interface{} {
	result := make([]interface{}, len(data))
	for k, v := range data {
		index, _ := strconv.Atoi(k)

		// Recursively check if the value is also a map that needs conversion
		switch nestedVal := v.(type) {
		case map[string]interface{}:
			if isNumericKeys(nestedVal) {
				result[index] = mapToArray(nestedVal) // Recursively convert nested map
			} else {
				result[index] = nestedVal // Keep as is if not numeric-keyed map
			}
		default:
			result[index] = nestedVal // Regular value, just assign it
		}
	}
	return result
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

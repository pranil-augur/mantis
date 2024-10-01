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

	inputMap, err := createInputMap(value, globalVars)
	if err != nil {
		return nil, err
	}

	injectedNode := astutil.Apply(f, nil, func(c astutil.Cursor) bool {
		n := c.Node()

		switch x := n.(type) {
		case *ast.Field:
			for _, attr := range x.Attrs {
				if strings.HasPrefix(attr.Text, "@runinject") {
					varName := parseRunInjectAttr(attr.Text)
					if val, ok := inputMap[varName]; ok {
						x.Value = createASTNodeForValue(val)
					}
				}
			}
		}
		return true
	})

	return injectedNode.(ast.Expr), nil
}

func createInputMap(value cue.Value, globalVars map[string]interface{}) (map[string]interface{}, error) {
	inputMap := make(map[string]interface{})
	inputsValue := value.LookupPath(cue.ParsePath("inputs"))

	if !inputsValue.Exists() {
		return inputMap, nil
	}

	inputsList, err := inputsValue.List()
	if err != nil {
		return nil, fmt.Errorf("failed to parse inputs: %v", err)
	}

	for inputsList.Next() {
		input := inputsList.Value()
		alias, err := input.LookupPath(cue.ParsePath("alias")).String()
		if err != nil {
			return nil, fmt.Errorf("failed to get alias for input: %v", err)
		}

		valueField := input.LookupPath(cue.ParsePath("value"))
		var valuePaths []string

		switch valueField.Kind() {
		case cue.StringKind:
			valuePath, err := valueField.String()
			if err != nil {
				return nil, fmt.Errorf("failed to get value path for input %s: %v", alias, err)
			}
			valuePaths = []string{valuePath}
		case cue.ListKind:
			iter, err := valueField.List()
			if err != nil {
				return nil, fmt.Errorf("failed to iterate over value paths for input %s: %v", alias, err)
			}
			for iter.Next() {
				path, err := iter.Value().String()
				if err != nil {
					return nil, fmt.Errorf("failed to get value path from list for input %s: %v", alias, err)
				}
				valuePaths = append(valuePaths, path)
			}
		default:
			return nil, fmt.Errorf("unexpected value type for input %s: %v", alias, valueField.Kind())
		}

		var inputValue interface{}
		for _, path := range valuePaths {
			if val, ok := lookupNestedValue(globalVars, path); ok {
				if len(valuePaths) == 1 {
					inputValue = val
				} else {
					if inputValue == nil {
						inputValue = make([]interface{}, 0)
					}
					inputValue = append(inputValue.([]interface{}), val)
				}
			} else {
				return nil, fmt.Errorf("failed to resolve input %s: %s", alias, path)
			}
		}

		inputMap[alias] = inputValue
	}

	return inputMap, nil
}

func parseRunInjectAttr(attrText string) string {
	attrText = strings.TrimPrefix(attrText, "@runinject(")
	attrText = strings.TrimSuffix(attrText, ")")
	return strings.Trim(attrText, "\"")
}

func lookupNestedValue(m map[string]interface{}, key string) (interface{}, bool) {
	parts := strings.Split(key, ".")
	current := m

	for i, part := range parts {
		if i == len(parts)-1 {
			// We've reached the final part, return the value
			if val, ok := current[part]; ok {
				return val, true
			}
			return nil, false
		}

		if val, ok := current[part]; ok {
			if nextMap, isMap := val.(map[string]interface{}); isMap {
				current = nextMap
			} else {
				// We've hit a non-map value before the end of the path
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return nil, false
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
		return &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(fmt.Sprintf("%v", v))}
	}
}

func updateGlobalVars(ctx *flowctx.Context, bt *task.BaseTask, value cue.Value) {
	taskPath := bt.ID
	outputsValue := bt.Final.LookupPath(cue.ParsePath("outputs"))
	outValue := value.LookupPath(cue.ParsePath("out"))

	// Parse outValue into a Go object
	var outData interface{}
	if err := outValue.Decode(&outData); err != nil {
		// Instead of returning immediately, let's try to work with the partial data
		fmt.Printf("Error decoding outValue: %v\n", err)
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
				processOutput(ctx, taskPath, alias, jqPath, outData)
			}
		default:
			fmt.Printf("Unexpected outputs kind: %v\n", outputsValue.Kind())
		}
	} else {
		fmt.Println("No outputs defined for this task")
	}
}

func processOutput(ctx *flowctx.Context, taskPath, alias, jqPath string, outData interface{}) {
	if actualValue, ok := queryJQ(outData, jqPath); ok {
		fullPath := fmt.Sprintf("tasks.%s.outputs.%s", taskPath, alias)
		setNestedValue(ctx.GlobalVars, fullPath, actualValue)
	} else {
		fmt.Printf("Value not found for output: %s at path: %s\n", alias, jqPath)
	}
}

func queryJQ(data interface{}, jqPath string) (interface{}, bool) {
	query, err := gojq.Parse(jqPath)
	if err != nil {
		fmt.Printf("Error parsing JQ query (%s): %v\n", jqPath, err)
		return nil, false
	}

	iter := query.Run(data)
	for {
		result, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := result.(error); isErr {
			fmt.Printf("Error during JQ query execution: %v\n", err)
			continue
		}
		// Found a non-error result
		return result, true
	}

	return nil, false
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

func fetchActualValue(value cue.Value, path []string) cue.Value {
	current := value
	for _, part := range path {
		current = current.LookupPath(cue.ParsePath(part))
		if !current.Exists() {
			fmt.Printf("Path part not found: %s\n, looking for key with quotes", part)
			current = value.LookupPath(cue.ParsePath(fmt.Sprintf("%q", cue.ParsePath(part))))
		}

		if !current.Exists() {
			fmt.Printf("Not found with double quotes. Looking for backticks %s\n", part)
			current = value.LookupPath(cue.ParsePath(fmt.Sprintf("`%s`", part)))
		}

		if !current.Exists() {
			fmt.Printf("Path part not found: %s\n", part)
			return cue.Value{}
		}
	}
	return current
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

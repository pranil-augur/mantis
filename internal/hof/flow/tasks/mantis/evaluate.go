/* Copyright 2024 Augur AI
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 */

package mantis

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/token"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type LocalEvaluator struct{}

func NewLocalEvaluator(val cue.Value) (hofcontext.Runner, error) {
	return &LocalEvaluator{}, nil
}

func parseRunInjectAttr(attrText string) string {
	attrText = strings.TrimPrefix(attrText, "@var(")
	attrText = strings.TrimSuffix(attrText, ")")
	return strings.Trim(attrText, "\"")
}

// @arr(var, index)
func parseArrayInjectAttr(attrText string) (string, int) {
	// Remove @arr( prefix and trailing )
	attrText = strings.TrimPrefix(attrText, "@arr(")
	attrText = strings.TrimSuffix(attrText, ")")

	// Split by comma, allowing for any amount of whitespace
	parts := strings.SplitN(attrText, ",", 2)
	if len(parts) != 2 {
		fmt.Printf("warning: Invalid @arr attribute format: %s\n", attrText)
		return "", 0
	}

	// Trim whitespace and quotes from variable name
	varName := strings.Trim(parts[0], " \t\"")

	// Trim whitespace and quotes from index, then parse
	indexStr := strings.Trim(parts[1], " \t\"")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		fmt.Printf("warning: Error parsing array index: %v\n", err)
		return "", 0
	}

	return varName, index
}

func createASTNodeForValue(val interface{}) ast.Expr {
	switch v := val.(type) {
	case cue.Value:
		return v.Syntax(cue.Final()).(ast.Expr)
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

func injectVariables(taskId string, value cue.Value, globalVars *sync.Map) (ast.Expr, error) {
	if globalVars == nil {
		return nil, fmt.Errorf("globalVars is nil")
	}

	f := value.Syntax(cue.Final())
	expr, ok := f.(ast.Expr)
	if !ok {
		return nil, fmt.Errorf("failed to convert value to ast.Expr for task %s", taskId)
	}

	// Check if the expression is valid before proceeding
	if expr == nil {
		return nil, fmt.Errorf("invalid or missing configuration for task %s", taskId)
	}

	// Process @preinject attributes before @runinject
	injectedNode := astutil.Apply(f, nil, func(c astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.Field:
			for _, attr := range x.Attrs {
				if strings.HasPrefix(attr.Text, "@var") {
					varName := parseRunInjectAttr(attr.Text)
					if val, ok := globalVars.Load(varName); ok {
						x.Value = createASTNodeForValue(val)
					}
				} else if strings.HasPrefix(attr.Text, "@arr") {
					varName, index := parseArrayInjectAttr(attr.Text)
					if val, ok := globalVars.Load(varName); ok {
						tempVal := createASTNodeForValue(val)
						if listLit, ok := tempVal.(*ast.ListLit); ok && index < len(listLit.Elts) {
							x.Value = listLit.Elts[index]
						}
					}
				}
			}
		}
		return true
	})

	return injectedNode.(ast.Expr), nil
}

// Run processes locals and dynamically evaluates expressions
func (T *LocalEvaluator) Run(ctx *hofcontext.Context) (interface{}, error) {
	v := ctx.Value
	if !ctx.Apply {
		return v, nil
	}
	ferr := func() error {
		ctx.CUELock.Lock()
		defer ctx.CUELock.Unlock()

		exports := v.LookupPath(cue.ParsePath("exports"))
		iter, _ := exports.List()

		// Print all global vars
		fmt.Println("Global Variables in evaluate.go:")
		ctx.GlobalVars.Range(func(key, value interface{}) bool {
			fmt.Printf("%v: %v\n", key, value)
			return true
		})

		// Create a new CUE context
		cueCtx := cuecontext.New()
		for iter.Next() {
			cueExpression := iter.Value().LookupPath(cue.ParsePath("cueexpr"))
			if !cueExpression.Exists() {
				return fmt.Errorf("path 'cueexpr' not found in CUE file")
			}
			varVal := iter.Value().LookupPath(cue.ParsePath("var"))
			varName, err := varVal.String()
			if err != nil {
				return err
			}

			exprStr, err := cueExpression.String()
			if err != nil {
				return err
			}

			// Create a temporary CUE file with the expression and necessary imports
			tmpFile, err := createTempCueFile(exprStr)
			if err != nil {
				return fmt.Errorf("failed to create temporary CUE file: %w", err)
			}
			defer os.Remove(tmpFile)

			// Load the CUE file
			instances := load.Instances([]string{tmpFile}, nil)
			if len(instances) == 0 {
				return fmt.Errorf("no instances loaded")
			}

			// Build the CUE value
			value := cueCtx.BuildInstance(instances[0])
			if value.Err() != nil {
				return fmt.Errorf("failed to build CUE instance: %w", value.Err())
			}
			injectedNode, err := injectVariables(ctx.BaseTask.ID, value, ctx.GlobalVars)
			if err != nil {
				return fmt.Errorf("failed to build CUE instance: %w", err)
			}
			newCueValue := ctx.CueContext.BuildExpr(injectedNode)

			result := newCueValue.LookupPath(cue.ParsePath("result"))

			// Serialize the CUE value to JSON
			jsonBytes, err := result.MarshalJSON()
			if err != nil {
				return fmt.Errorf("failed to marshal CUE value to JSON: %w", err)
			}

			// Create a new slice and copy the data
			persistentBytes := make([]byte, len(jsonBytes))
			copy(persistentBytes, jsonBytes)

			// Store a pointer to the persistent JSON bytes in GlobalVars
			ctx.GlobalVars.Store(varName, &persistentBytes)

			fmt.Printf("Storing pointer to serialized CUE value in GlobalVars - %s: %p\n", varName, &persistentBytes)
		}
		return nil
	}()

	if ferr != nil {
		return nil, ferr
	}

	// Return updated CUE context with evaluated locals
	return ctx.Value, nil
}

func createTempCueFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "evaluate*.cue")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

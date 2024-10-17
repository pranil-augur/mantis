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
	"regexp"
	"strconv"
	"strings"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/token"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/zclconf/go-cty/cty"
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
		// Check if n is a field node
		if field, ok := n.(*ast.Field); ok {
			for _, attr := range field.Attrs {
				if strings.HasPrefix(attr.Text, "@var") {
					varName := parseRunInjectAttr(attr.Text)
					if val, ok := globalVars.Load(varName); ok {
						field.Value = createASTNodeForValue(val)
					}
				} else if strings.HasPrefix(attr.Text, "@arr") {
					varName, index := parseArrayInjectAttr(attr.Text)
					if val, ok := globalVars.Load(varName); ok {
						tempVal := createASTNodeForValue(val)
						if listLit, ok := tempVal.(*ast.ListLit); ok && index < len(listLit.Elts) {
							field.Value = listLit.Elts[index]
						}
					}
				}
			}
		}
		return true
	})
	return injectedNode.(ast.Expr), nil
}
func FormatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v) // Quote string values
	case []interface{}:
		// Handle arrays of interface{}
		formattedElements := make([]string, len(v))
		for i, elem := range v {
			formattedElements[i] = FormatValue(elem) // Recursively format each element
		}
		return fmt.Sprintf("[%s]", strings.Join(formattedElements, ", "))
	case int, float64, bool:
		// Return numbers and booleans without quotes
		return fmt.Sprintf("%v", v)
	default:
		// Handle any other types with default formatting
		return fmt.Sprintf("%v", v)
	}
}
func PopulateTemplate(template string, vars *sync.Map) (string, error) {
	// Define a regex pattern to match @var(id_here)
	re := regexp.MustCompile(`@var\((\w+)\)`)

	// Function to replace each match with the corresponding value from the sync.Map
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract the variable name using capturing groups
		matches := re.FindStringSubmatch(match)
		if len(matches) > 1 {
			varName := matches[1] // This will be the captured variable name

			// Load the value from sync.Map
			if value, exists := vars.Load(varName); exists {
				// Use the FormatValue function to handle the formatting
				return FormatValue(value)
			}
		}
		return match // If the variable doesn't exist, return the original match
	})
	return result, nil
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

		for iter.Next() {
			cueExpression := iter.Value().LookupPath(cue.ParsePath("cueexpr"))

			if !cueExpression.Exists() {
				return fmt.Errorf("path 'cueexpr' not found in CUE file")
			}

			exprStr, err := cueExpression.String()
			if err != nil {
				return err
			}

			populatedCueExpr, err := PopulateTemplate(exprStr, ctx.GlobalVars)
			if err != nil {
				return fmt.Errorf("failed to update variables in CUE file: %w", err)
			}
			// Create a temporary CUE file with the expression and necessary imports
			tmpFile, err := createTempCueFile(populatedCueExpr)
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
			value := ctx.CueContext.BuildInstance(instances[0])
			if value.Err() != nil {
				return fmt.Errorf("failed to build CUE instance: %w", value.Err())
			}

			result := value.LookupPath(cue.ParsePath("result"))
			v = v.FillPath(cue.ParsePath("out"), result)
			ctx.Value = v

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

func convertCtyToGo(input *sync.Map) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var conversionError error

	input.Range(func(key, value interface{}) bool {
		convertedValue, err := convertValue(value)
		if err != nil {
			conversionError = fmt.Errorf("error converting key '%v': %w", key, err)
			return false // Stop iteration on error
		}

		// Assuming the key is a string, if not, you may need to convert it
		strKey, ok := key.(string)
		if !ok {
			conversionError = fmt.Errorf("key is not a string: %v", key)
			return false // Stop iteration
		}

		result[strKey] = convertedValue
		return true // Continue iteration
	})

	if conversionError != nil {
		return nil, conversionError
	}

	return result, nil
}

func convertValue(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case cty.Value:
		return ctyValueToGo(v)
	case map[string]cty.Value:
		return convertCtyValueMap(v)
	case sync.Map:
		return convertCtyToGo(&v)
	case *sync.Map:
		return convertCtyToGo(v)
	case []interface{}:
		return convertSlice(v)
	default:
		// If it's not a recognized type, return it as-is
		return v, nil
	}
}

func convertCtyValueMap(input map[string]cty.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range input {
		convertedValue, err := ctyValueToGo(value)
		if err != nil {
			return nil, fmt.Errorf("error converting key '%s': %w", key, err)
		}
		result[key] = convertedValue
	}
	return result, nil
}

func convertSlice(slice []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		convertedValue, err := convertValue(v)
		if err != nil {
			return nil, fmt.Errorf("error converting slice element at index %d: %w", i, err)
		}
		result[i] = convertedValue
	}
	return result, nil
}

func ctyValueToGo(v cty.Value) (interface{}, error) {
	if v.IsNull() {
		return nil, nil
	}

	switch {
	case v.Type() == cty.String:
		return v.AsString(), nil
	case v.Type() == cty.Number:
		return v.AsBigFloat(), nil
	case v.Type() == cty.Bool:
		return v.True(), nil
	case v.Type().IsListType() || v.Type().IsTupleType():
		return ctyListToSlice(v)
	case v.Type().IsMapType() || v.Type().IsObjectType():
		return ctyMapToMap(v)
	case v.Type().IsSetType():
		return ctySetToSlice(v)
	default:
		// Instead of returning an error, let's return the string representation
		return v.GoString(), nil
	}
}

func ctySetToSlice(v cty.Value) ([]interface{}, error) {
	if !v.Type().IsSetType() {
		return nil, fmt.Errorf("not a set type")
	}

	result := make([]interface{}, 0, v.LengthInt())
	for it := v.ElementIterator(); it.Next(); {
		_, ev := it.Element()
		goValue, err := ctyValueToGo(ev)
		if err != nil {
			return nil, fmt.Errorf("error converting set element: %w", err)
		}
		result = append(result, goValue)
	}

	return result, nil
}

func ctyListToSlice(v cty.Value) ([]interface{}, error) {
	length := v.LengthInt()
	result := make([]interface{}, length)

	for i := 0; i < length; i++ {
		element := v.Index(cty.NumberIntVal(int64(i)))
		goValue, err := ctyValueToGo(element)
		if err != nil {
			return nil, fmt.Errorf("error converting list element at index %d: %w", i, err)
		}
		result[i] = goValue
	}

	return result, nil
}

func ctyMapToMap(v cty.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for it := v.ElementIterator(); it.Next(); {
		key, value := it.Element()
		keyString, err := ctyValueToGo(key)
		if err != nil {
			return nil, fmt.Errorf("error converting map key: %w", err)
		}

		keyStr, ok := keyString.(string)
		if !ok {
			return nil, fmt.Errorf("map key is not a string: %v", keyString)
		}

		goValue, err := ctyValueToGo(value)
		if err != nil {
			return nil, fmt.Errorf("error converting map value for key '%s': %w", keyStr, err)
		}

		result[keyStr] = goValue
	}

	return result, nil
}

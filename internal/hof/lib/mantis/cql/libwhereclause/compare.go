package libwhereclause

import (
	"fmt"
	"regexp"

	"cuelang.org/go/cue"
)

func compareEqual(value cue.Value, expected any) bool {
	switch v := expected.(type) {
	case string:
		str, err := value.String()
		return err == nil && str == v
	case int:
		num, err := value.Int64()
		return err == nil && int(num) == v
	case int64:
		num, err := value.Int64()
		return err == nil && num == v
	case float64:
		num, err := value.Float64()
		return err == nil && num == v
	case bool:
		b, err := value.Bool()
		return err == nil && b == v
	case []interface{}: // For handling slices/arrays
		list, err := value.List()
		if err != nil {
			return false
		}
		// Compare each element
		i := 0
		for list.Next() {
			if i >= len(v) || !compareEqual(list.Value(), v[i]) {
				return false
			}
			i++
		}
		return i == len(v)
	case map[string]interface{}: // For handling nested objects
		obj := value.Eval()
		if obj.Err() != nil {
			return false
		}
		iter, _ := obj.Fields()
		for iter.Next() {
			expectedVal, ok := v[iter.Label()]
			if !ok || !compareEqual(iter.Value(), expectedVal) {
				return false
			}
		}
		return true
	}

	// Print the type and value of expected for debugging
	fmt.Printf("Type: %T, Value: %v\n", expected, expected)
	fmt.Printf("Type: %T, Value: %v\n", value, value)
	return false
}

func compareRegex(value cue.Value, pattern any) bool {
	str, err := value.String()
	if err != nil {
		return false
	}
	patternStr, ok := pattern.(string)
	if !ok {
		return false
	}
	matched, err := regexp.MatchString(patternStr, str)
	return err == nil && matched
}

func compareIn(value cue.Value, list any) bool {
	// If list is a cue.Value, handle it as a CUE list
	if listValue, ok := list.(cue.Value); ok {
		iter, err := listValue.List()
		if err != nil {
			return false
		}
		// Check if value matches any element in the list
		for iter.Next() {
			if value.Equals(iter.Value()) {
				return true
			}
		}
		return false
	}
	return false
}

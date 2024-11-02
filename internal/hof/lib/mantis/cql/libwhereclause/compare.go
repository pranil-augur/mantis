package libwhereclause

import (
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
	case bool:
		b, err := value.Bool()
		return err == nil && b == v
	}
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
	// TODO: Implement array/list membership check
	return false
}

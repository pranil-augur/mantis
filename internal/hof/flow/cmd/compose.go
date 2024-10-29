package cmd

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

func ParseComposeFile(filePath string) ([]string, error) {
	ctx := cuecontext.New()

	// Load the CUE file
	instances := load.Instances([]string{filePath}, nil)
	if len(instances) == 0 {
		return nil, fmt.Errorf("no CUE instances found")
	}

	// Build the CUE value
	value := ctx.BuildInstance(instances[0])
	if value.Err() != nil {
		return nil, value.Err()
	}

	// Look for the commands field specifically
	commands := value.LookupPath(cue.ParsePath("commands"))
	if !commands.Exists() {
		return nil, fmt.Errorf("no commands field found in CUE file")
	}

	// Extract the commands in order
	var result []string
	iter, err := commands.Fields()
	if err != nil {
		return nil, err
	}

	for iter.Next() {
		cmdArray, err := iter.Value().List()
		if err != nil {
			return nil, fmt.Errorf("command %q is not a list: %v", iter.Label(), err)
		}

		// Build the command string from the array
		var cmdParts []string
		for cmdArray.Next() {
			str, err := cmdArray.Value().String()
			if err != nil {
				return nil, fmt.Errorf("invalid command part: %v", err)
			}
			cmdParts = append(cmdParts, str)
		}

		result = append(result, strings.Join(cmdParts, " "))
	}

	return result, nil
}

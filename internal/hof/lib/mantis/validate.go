/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package mantis

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/token"
)

func Validate(dir string) error {
	// Create a new CUE context
	ctx := cuecontext.New()

	// Load CUE files from the specified directory
	cfg := &load.Config{
		Dir: dir,
	}
	instances := load.Instances([]string{"."}, cfg)

	if len(instances) == 0 {
		return fmt.Errorf("no CUE files found in directory: %s", dir)
	}

	var validationErrors []errors.Error // Change this to store CUE errors directly

	for _, inst := range instances {
		fmt.Printf("Validating file: %s\n", inst.Dir)

		if inst.Err != nil {
			fmt.Printf("Warning: Error loading %s: %v\n", inst.Dir, inst.Err)
			validationErrors = append(validationErrors, inst.Err)
			continue
		}

		value := ctx.BuildInstance(inst)

		// Validate with concrete set to true
		opt := []cue.Option{
			cue.Attributes(true),
			cue.Definitions(true),
			cue.Hidden(true),
			cue.Concrete(true),
		}

		err := value.Validate(opt...)
		if err != nil {
			fmt.Printf("Validation errors in %s\n", inst.Dir)
			if cueerr, ok := err.(errors.Error); ok {
				validationErrors = append(validationErrors, cueerr)
			} else {
				// This should rarely happen, but just in case
				validationErrors = append(validationErrors, errors.Newf(token.NoPos, err.Error()))
			}
		} else {
			fmt.Printf("Validation successful for %s\n", inst.Dir)
		}
	}

	// If there were any errors, print them in a more readable format
	if len(validationErrors) > 0 {
		fmt.Fprintln(os.Stderr, "Validation failed. The following errors were found:")
		for i, err := range validationErrors {
			fmt.Fprintf(os.Stderr, "\nError %d:\n", i+1)
			printDetailedError(err)
		}
		return fmt.Errorf("validation failed with %d error(s)", len(validationErrors))
	}

	fmt.Println("Validation successful! All CUE files in the directory are valid.")
	return nil
}

func printDetailedError(err errors.Error) {
	errs := errors.Errors(err)
	for i, e := range errs {
		fmt.Fprintf(os.Stderr, "Error %d:\n", i+1)
		fmt.Fprintln(os.Stderr, errors.Details(e, nil))
		printErrorContext(e)
		fmt.Fprintln(os.Stderr)
	}
}

func printErrorContext(err errors.Error) {
	for _, f := range err.InputPositions() {
		if f.Filename() == "" {
			continue
		}
		content, err := os.ReadFile(f.Filename())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", f.Filename(), err)
			continue
		}

		lines := strings.Split(string(content), "\n")
		line := f.Line()
		column := f.Column()
		if line > 0 && line <= len(lines) {
			fmt.Fprintf(os.Stderr, "\nFile: %s\n", f.Filename())

			// Print a few lines before the error
			start := max(0, line-3)
			for i := start; i < line-1; i++ {
				fmt.Fprintf(os.Stderr, "%5d | %s\n", i+1, lines[i])
			}

			// Print the error line
			errorLine := lines[line-1]
			fmt.Fprintf(os.Stderr, "%5d | %s\n", line, errorLine)

			// Highlight the specific character
			pointerLine := strings.Repeat(" ", column+5) + "^"
			fmt.Fprintln(os.Stderr, pointerLine)

			// Print the character code if it's a non-printable character
			if column < len(errorLine) {
				char := errorLine[column]
				if char < 32 || char > 126 {
					fmt.Fprintf(os.Stderr, "%sU+%04X\n", strings.Repeat(" ", column+5), char)
				}
			}

			// Print a few lines after the error
			end := min(len(lines), line+2)
			for i := line; i < end; i++ {
				fmt.Fprintf(os.Stderr, "%5d | %s\n", i+1, lines[i])
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

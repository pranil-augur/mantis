/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package test

import (
	"fmt"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/cuetils"
)

func RunTestFromArgsFlags(args []string, cmdflags flags.TestFlagpole) error {

	verbose := flags.RootPflags.Verbose
	cueFiles, extraArgs := args, []string{}

	// split args at "--"
	pos := -1
	for i, arg := range args {
		if arg == "--" {
			pos = i
			break
		}
	}
	if pos >= 0 {
		cueFiles, extraArgs = args[0:pos], args[pos+1:]
		if len(extraArgs) > 0 {
			fmt.Println("using extra args:", extraArgs)
		}
	}

	// Loadup our Cue files
	crt, err := cuetils.CueRuntimeFromEntrypointsAndFlags(cueFiles)
	if err != nil {
		return err
	}

	// Get test suites from top level
	suites, err := getValueTestSuites(crt.CueContext, crt.CueValue, cmdflags.Suite)
	if err != nil {
		return err
	}

	// find tests in suites
	for s, suite := range suites {
		ts, err := getValueTestSuiteTesters(crt.CueContext, suite.Value, cmdflags.Tester)
		if err != nil {
			return err
		}
		// make sure to write to original
		suites[s].Tests = ts
	}

	// Is the user only looking for information
	if cmdflags.List {
		printTests(suites, false)
		return nil
	}

	// Run all of our suites
	_, err = RunSuites(suites, verbose)

	// Print our final tests and stats
	fmt.Printf("\n\n\n======= FINAL RESULTS ======\n")
	printTests(suites, true)
	fmt.Println("============================")

	// Finally, check for errors and exit appropriately
	if err != nil {
		return err
	}

	for _, s := range suites {
		if len(s.Errors) > 0 {
			return fmt.Errorf("\nErrors during testing")
		}
	}

	return nil
}

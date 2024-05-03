/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package main

import (
	"fmt"
	"os"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	runner "github.com/opentofu/opentofu/internal/hof/flow/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "cuestack"}
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

var runCmd = &cobra.Command{
	Use:   "run [path]",
	Short: "Run a cue flow from a file or directory",
	Long:  `Run a cue flow from a file or directory specified by the path argument.`,
	Args:  cobra.ExactArgs(1),
	Run:   runFlowFromFileOrDir,
}

func runFlowFromFileOrDir(cmd *cobra.Command, args []string) {
	// Assuming args[0] is the path to the file or directory containing the flow
	flowPath := args[0]

	// Prepare the runtime with default flags
	rflags := flags.RootPflagpole{}
	cflags := flags.FlowPflagpole{}

	// Convert the flowPath into a format that can be passed to Run
	// Assuming Run can take the flowPath directly as part of args
	argsForRun := []string{flowPath}

	// Call Run from run.go
	err := runner.Run(argsForRun, rflags, cflags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running flow: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Flow completed successfully")
}

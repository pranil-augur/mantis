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

var rootCmd = &cobra.Command{Use: "mantis"}
var runCmd = &cobra.Command{
	Use:   "run [path]",
	Short: "Run a cue flow from a file or directory",
	Long:  `Run a cue flow from a file or directory specified by the path argument.`,
	Args:  cobra.ExactArgs(1),
	Run:   runFlowFromFileOrDir,
}

var genCmd = &cobra.Command{
	Use:   "gen <target directory> <package name>",
	Short: "Generate scaffolding for a new cue module",
	Long:  `Generate scaffolding for a new cue module in the specified target directory with the given package name.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runner.Gen(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating scaffolding: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Scaffolding generated successfully.")
	},
}

var rflags flags.RootPflagpole

func init() {
	// Initialize flags using the function from root.go
	// flags.SetupRootPflags(rootCmd.PersistentFlags(), &rflags)
	rootCmd.PersistentFlags().StringArrayVarP(&(rflags.Tags), "tags", "t", nil, "@tags() to be injected into CUE code")
	rootCmd.PersistentFlags().BoolVarP(&(rflags.InjectEnv), "inject-env", "V", false, "inject all ENV VARs as default tag vars")
	rootCmd.PersistentFlags().BoolVarP(&rflags.Preview, "preview", "P", false, "preview the changes to the state")
	rootCmd.PersistentFlags().BoolVarP(&rflags.Apply, "apply", "A", false, "apply the proposed state")
	rootCmd.PersistentFlags().BoolVarP(&rflags.Init, "init", "I", false, "init modules")
	rootCmd.PersistentFlags().BoolVarP(&rflags.Destroy, "destroy", "D", false, "destroy resources")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(genCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func runFlowFromFileOrDir(cmd *cobra.Command, args []string) {

	// Assuming args[0] is the path to the file or directory containing the flow
	flowPath := args[0]

	// Prepare the runtime with initialized flags
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

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
	"io"
	"os"
	"path/filepath"
	"strings"

	"net/http"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"github.com/abiosoft/hcl2json"
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

var importCmd = &cobra.Command{
	Use:   "import <import directory> <output directory>",
	Short: "Import and convert Terraform Module/file to Cue Module/File",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run:   convertHCLtoCue,
}

var rflags flags.RootPflagpole

func init() {
	// Initialize flags using the function from root.go
	// flags.SetupRootPflags(rootCmd.PersistentFlags(), &rflags)
	rootCmd.PersistentFlags().StringArrayVarP(&(rflags.Tags), "tags", "t", nil, "@tags() to be injected into CUE code")
	rootCmd.PersistentFlags().BoolVarP(&(rflags.InjectEnv), "inject-env", "V", false, "inject all ENV VARs as default tag vars")
	rootCmd.PersistentFlags().BoolVarP(&rflags.Plan, "plan", "P", false, "plan the changes to the state")
	rootCmd.PersistentFlags().BoolVarP(&rflags.Gist, "gist", "G", false, "gist of changes")
	rootCmd.PersistentFlags().BoolVarP(&rflags.Apply, "apply", "A", false, "apply the proposed state")
	rootCmd.PersistentFlags().BoolVarP(&rflags.Init, "init", "I", false, "init modules")
	rootCmd.PersistentFlags().BoolVarP(&rflags.Destroy, "destroy", "D", false, "destroy resources")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(importCmd)
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
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func convertHCLtoCue(cmd *cobra.Command, args []string) {
	importPath := args[0]
	outputPath := ""

	if len(args) > 1 {
		outputPath = args[1]
	}

	var hclData []byte
	var err error

	if strings.HasPrefix(importPath, "http://") || strings.HasPrefix(importPath, "https://") {
		fmt.Println("Fetching TF File from URL...")
		hclData, err = downloadFile(importPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error downloading file from %q: %v\n", importPath, err)
			return
		}
	} else {

		fileInfo, err := os.Stat(importPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error accessing path %q: %v\n", importPath, err)
			return
		}

		if fileInfo.IsDir() {

			err := processModule(importPath, outputPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error processing module: %v\n", err)
			}
			return
		}

		hclData, err = os.ReadFile(importPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file %q: %v\n", importPath, err)
			return
		}
	}

	err = processHCLData(hclData, outputPath, importPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error processing HCL data: %v\n", err)
	}
}

func processHCLData(hclData []byte, outputPath, importPath string) error {
	converter := hcl2json.New(hclData)

	jsonOutput, err := converter.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert HCL to JSON: %v", err)
	}

	ctx := cuecontext.New()

	cueValue := ctx.CompileBytes(jsonOutput)
	if cueValue.Err() != nil {
		return fmt.Errorf("failed to compile JSON to CUE: %v", cueValue.Err())
	}

	opts := []cue.Option{
		cue.Concrete(true),
		cue.Definitions(true),
		cue.Optional(true),
		cue.Hidden(true),
		cue.Attributes(true),
		cue.Docs(true),
		cue.InlineImports(true),
		cue.ErrorsAsValues(true),
	}
	cueData, err := format.Node(cueValue.Syntax(opts...))
	if err != nil {
		return fmt.Errorf("failed to format CUE: %v", err)
	}

	if outputPath == "" {
		outputPath = filepath.Join(".", strings.TrimSuffix(filepath.Base(importPath), filepath.Ext(importPath))+".cue")
	} else if outputPath[len(outputPath)-1] == '/' {
		outputPath = filepath.Join(outputPath, strings.TrimSuffix(filepath.Base(importPath), filepath.Ext(importPath))+".cue")
	} else {
		outputPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".cue"
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	_, err = outputFile.Write(cueData)
	if err != nil {
		return fmt.Errorf("failed to write CUE to output file: %v", err)
	}

	fmt.Println("Conversion successful:", outputPath)
	return nil
}

func processModule(importPath, outputPath string) error {
	info, err := os.Stat(importPath)
	if err != nil {
		return fmt.Errorf("failed to stat import path: %v", err)
	}

	if info.IsDir() {

		return filepath.Walk(importPath, func(filePath string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fileInfo.IsDir() {
				return nil
			}

			if filepath.Ext(filePath) == ".tf" {

				relativePath, err := filepath.Rel(importPath, filePath)
				if err != nil {
					return fmt.Errorf("failed to determine relative path: %v", err)
				}

				outputFilePath := filepath.Join(outputPath, strings.TrimSuffix(relativePath, ".tf")+".cue")

				if err := os.MkdirAll(filepath.Dir(outputFilePath), os.ModePerm); err != nil {
					return fmt.Errorf("failed to create directories for output path: %v", err)
				}

				return processHCLFile(filePath, outputFilePath)
			}
			return nil
		})
	} else {

		if filepath.Ext(importPath) == ".tf" {
			return processHCLFile(importPath, outputPath)
		}
		return fmt.Errorf("import path is not a valid .tf file or directory")
	}
}

func processHCLFile(importFilePath, outputFilePath string) error {
	terraformFile, err := os.Open(importFilePath)
	if err != nil {
		return fmt.Errorf("failed to open terraform file: %v", err)
	}
	defer terraformFile.Close()

	hclData, err := io.ReadAll(terraformFile)
	if err != nil {
		return fmt.Errorf("failed to read terraform file: %v", err)
	}

	converter := hcl2json.New(hclData)
	jsonOutput, err := converter.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert HCL to JSON: %v", err)
	}

	ctx := cuecontext.New()
	cueValue := ctx.CompileBytes(jsonOutput)
	if cueValue.Err() != nil {
		return fmt.Errorf("failed to compile JSON to CUE: %v", cueValue.Err())
	}

	opts := []cue.Option{
		cue.Concrete(true),
		cue.Definitions(true),
		cue.Optional(true),
		cue.Hidden(true),
		cue.Attributes(true),
		cue.Docs(true),
		cue.InlineImports(true),
		cue.ErrorsAsValues(true),
	}

	cueData, err := format.Node(cueValue.Syntax(opts...))
	if err != nil {
		return fmt.Errorf("failed to format CUE: %v", err)
	}

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	_, err = outputFile.Write(cueData)
	if err != nil {
		return fmt.Errorf("failed to write CUE to output file: %v", err)
	}

	fmt.Printf("Converted %v to %v\n", importFilePath, outputFilePath)
	return nil
}

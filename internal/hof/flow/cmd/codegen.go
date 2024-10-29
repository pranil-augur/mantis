/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */
package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/opentofu/opentofu/internal/hof/lib/codegen"
	"github.com/opentofu/opentofu/internal/hof/lib/codegen/types"
	"github.com/opentofu/opentofu/internal/hof/lib/mantis"
	"github.com/spf13/cobra"
)

// Codegen provides the main interface for code generation.
type Codegen struct {
	AIGen             *codegen.AiGen
	SystemPrompt      string
	SystemPromptPath  string
	UserPrompt        string
	CodeDir           string
	MantisParams      string
	MaxAttempts       int
	CurrentAttempt    int
	Context           string
	AdditionalContext string
}

// New constructs a new Codegen object with the path to a configuration file.
func New(confPath, systemPromptPath, codeDir, userPrompt string) (*Codegen, error) {
	aigen, err := codegen.New(confPath)
	if err != nil {
		return nil, fmt.Errorf("failed initializing Codegen: %w", err)
	}

	systemPrompt, err := loadPromptFromPath(systemPromptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load system prompt: %w", err)
	}

	return &Codegen{
		AIGen:             aigen,
		MaxAttempts:       5, // Default value, can be customized
		CurrentAttempt:    0,
		SystemPrompt:      systemPrompt,
		SystemPromptPath:  systemPromptPath,
		UserPrompt:        userPrompt,
		CodeDir:           codeDir,
		AdditionalContext: "",
	}, nil
}

// Run executes the code generation process based on the given task description.
func (c *Codegen) Run() error {
	ctx := context.Background()
	fmt.Printf("Starting agent with task: %s\n", c.UserPrompt)

	chat, err := c.AIGen.Chat(ctx, "", "")
	if err != nil {
		return fmt.Errorf("failed to initialize chat: %w", err)
	}

	var lastOutput string

	for c.CurrentAttempt < c.MaxAttempts {
		c.CurrentAttempt++

		// Introduce a small delay to prevent overwhelming the LLM
		time.Sleep(2 * time.Second)
		fmt.Printf("Attempt %d of %d\n", c.CurrentAttempt, c.MaxAttempts)

		// 1. Read code files and build context, including the last output
		if err := c.buildContext(lastOutput); err != nil {
			return fmt.Errorf("failed to build context: %w", err)
		}

		// 2. Generate code
		generatedCode, err := c.generateCode(ctx, chat, c.UserPrompt)
		if err != nil {
			return fmt.Errorf("failed to generate code: %w", err)
		}
		filePath, err := c.writeCode(generatedCode)
		// 3. Write generated code to file
		if err != nil {
			return fmt.Errorf("failed to write code: %w", err)
		}

		// 4. Validate generated code using Mantis
		validationOutput, err := c.runValidate()
		if err != nil {
			fmt.Printf("Validation failed: %v\n", err)
		}

		// 5. Analyze validation output
		if c.isCodeValid(validationOutput) {
			pwd, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			prompt := promptui.Select{
				Label: "Do you want me to run the code?",
				Items: []string{"Yes", "No"},
			}
			_, result, err := prompt.Run()
			if err != nil {
				return fmt.Errorf(" Prompt failed: %v", err)
			}

			fmt.Println("Code validation successful!")

			if result == "Yes" {

				changeDirectoryToScripts(pwd, filePath)
				fileName := getFileName(filePath)

				if err := runInit(fileName); err != nil {
					return fmt.Errorf("run --init failed: %w", err)
				}
				if err := runPlan(fileName); err != nil {
					return fmt.Errorf("run --plan failed: %w", err)
				}
				if err := runApply(fileName); err != nil {
					return fmt.Errorf("run --apply failed: %w", err)
				}
			}
			return nil
		}

		// 6. If code is not valid, prepare feedback for regeneration
		regenerationPrompt := c.prepareRegenerationPrompt(c.UserPrompt, generatedCode, validationOutput)

		// 7. Update context with validation results for next iteration
		lastOutput = fmt.Sprintf("Generated code:\n%s\n\nValidation output:\n%s", generatedCode, validationOutput)

		// The loop will continue with the updated context and regeneration prompt
		c.UserPrompt = regenerationPrompt
	}

	return fmt.Errorf("max attempts reached without successful code generation and validation")
}

func (c *Codegen) buildContext(lastOutput string) error {
	var context strings.Builder
	err := filepath.Walk(c.CodeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".cue" {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			context.WriteString(fmt.Sprintf("File: %s\n%s\n\n", path, string(content)))
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Add the last output to the context
	if lastOutput != "" {
		context.WriteString(fmt.Sprintf("Last Output:\n%s\n\n", lastOutput))
	}

	c.Context = context.String()
	return nil
}

func (c *Codegen) generateCode(ctx context.Context, chat types.Conversation, prompt string) (string, error) {
	combinedPrompt := fmt.Sprintf("System: %s\n\nUser: %s\n\nGiven the following context and instructions, generate the necessary code:\n\nContext:\n%s\n\nAdditional Context:\n%s\n\nInstructions:\n%s",
		c.SystemPrompt, prompt, c.Context, c.AdditionalContext, prompt)

	// Open the log file in append mode
	logFile, err := os.OpenFile("codegen.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	// Create a new logger
	logger := log.New(logFile, "", log.LstdFlags)

	// Log the prompt
	logger.Printf("Attempt %d - Sending prompt to LLM:\n%s\n", c.CurrentAttempt, combinedPrompt)

	// Print to console as well
	fmt.Printf("Sending prompt to LLM (Attempt %d)\n", c.CurrentAttempt)

	response, err := chat.Send(ctx, combinedPrompt)
	if err != nil {
		return "", err
	}

	// Log the response
	logger.Printf("Attempt %d - LLM Response:\n%s\n", c.CurrentAttempt, response.FullOutput)

	return response.FullOutput, nil
}

func (c *Codegen) writeCode(code string) (string, error) {
	filename := "flow.tf.cue"
	fullPath := filepath.Join(c.CodeDir, filename)

	// Move the previous attempt's code if it exists
	// This wi
	if c.CurrentAttempt > 1 {
		prevAttemptPath := fmt.Sprintf("%s.attempt%d", fullPath, c.CurrentAttempt-1)
		if err := os.Rename(fullPath, prevAttemptPath); err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to move previous attempt: %w", err)
		}
	}

	if err := os.MkdirAll(c.CodeDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(fullPath, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	fmt.Printf("Generated code written to: %s\n", fullPath)
	return fullPath, nil
}

func (c *Codegen) isCodeValid(validationOutput string) bool {
	// Implement logic to determine if the code is valid based on the validation output
	// This is a simplified example; adjust according to your specific validation output format
	return !strings.Contains(validationOutput, "error")
}

func (c *Codegen) prepareRegenerationPrompt(originalPrompt, generatedCode, validationOutput string) string {
	return fmt.Sprintf(`
Original prompt: %s

Generated code:
%s

Validation issues:
%s

Please fix the issues identified in the validation output and regenerate the code. 
Please dont include any formatting in the response, just the code.`, originalPrompt, generatedCode, validationOutput)
}

func (c *Codegen) runValidate() (string, error) {
	var output strings.Builder
	err := mantis.Validate(c.CodeDir)
	if err != nil {
		output.WriteString(fmt.Sprintf("Validation failed: %v", err))
	} else {
		output.WriteString("Validation successful!")
	}
	return output.String(), err
}

type Action struct {
	Type    string
	Content string
}

// Helper function to load prompt from a path
func loadPromptFromPath(path string) (string, error) {
	// Check if the location is a URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		if _, err := url.ParseRequestURI(path); err == nil {
			resp, err := http.Get(path)
			if err != nil {
				return "", fmt.Errorf("failed to fetch prompt from URL: %w", err)
			}
			defer resp.Body.Close()
			content, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("failed to read prompt from URL response: %w", err)
			}
			return string(content), nil
		}
	}

	// It's a local path, check if it's a file or directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat prompt location: %w", err)
	}

	if fileInfo.IsDir() {
		var prompt strings.Builder
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".txt" {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				prompt.WriteString(string(content))
				prompt.WriteString("\n")
			}
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("failed to read prompt files from directory: %w", err)
		}
		return prompt.String(), nil
	} else {
		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read prompt file: %w", err)
		}
		return string(content), nil
	}
}

var runCmd = &cobra.Command{
	Use:   "run [filename]",
	Short: "Run a cue flow with specified flags",
	Long:  `Run a cue flow using the specified filename with --init, --plan, or --apply flags.`,
	Args:  cobra.ExactArgs(1),
}

func runInit(filename string) error {
	fmt.Printf("🚀 Running mantis run --init %v\n", filename)
	cmd := exec.Command("mantis", "run", "--init", filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running --init: %v\nOutput: %s", err, string(output))
	}
	fmt.Printf("Output of run --init:\n%s\n", string(output))
	return nil
}

func runPlan(filename string) error {
	fmt.Printf("📋 Running mantis run --plan %v \n", filename)
	fmt.Println()
	cmd := exec.Command("mantis", "run", "--plan", filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running --plan: %v\nOutput: %s", err, string(output))
	}
	fmt.Printf("Output of run --plan:\n%s\n", string(output))
	return nil
}
func runApply(filename string) error {
	fmt.Printf("💻 Running mantis run --apply %v \n", filename)
	cmd := exec.Command("mantis", "run", "--apply", filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running --apply: %v\nOutput: %s", err, string(output))
	}
	fmt.Printf("Output of run --apply:\n%s\n", string(output))
	return nil
}

func init() {
	runCmd.Run = func(cmd *cobra.Command, args []string) {
		filename := args[0]

		if err := runInit(filename); err != nil {
			fmt.Println(err)
			return
		}

	}
}
func getFileName(filePath string) string {
	return filepath.Base(filePath)
}
func changeDirectoryToScripts(pwd, filePath string) error {
	dirName := filepath.Dir(filePath)
	dirPath := filepath.Join(pwd, dirName)
	err := os.Chdir(dirPath)
	if err != nil {
		return fmt.Errorf("error changing directory: %w", err)
	}

	return nil
}

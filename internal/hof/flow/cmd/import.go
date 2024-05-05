package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ImportHCL(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: importHCL <source directory>")
	}
	sourceDir := args[0]

	// Check if the source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	// Find all HCL files in the directory
	hclFiles, err := filepath.Glob(filepath.Join(sourceDir, "*.hcl"))
	if err != nil {
		return fmt.Errorf("error finding HCL files: %v", err)
	}

	if len(hclFiles) == 0 {
		return fmt.Errorf("no HCL files found in directory: %s", sourceDir)
	}

	// Read and concatenate HCL files content
	var allHCLContent strings.Builder
	for _, file := range hclFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", file, err)
		}
		allHCLContent.WriteString(string(content) + "\n")
	}

	// Generate the prompt for OpenAI
	prompt := fmt.Sprintf("Convert the following HCL scripts into a Cuestack blueprint:\n%s", allHCLContent.String())

	// Send the prompt to OpenAI and get the response (mocked here)
	cuestackBlueprint, err := SendToOpenAI(prompt)
	if err != nil {
		return fmt.Errorf("error generating Cuestack blueprint from OpenAI: %v", err)
	}

	// Write the output to a .patch file
	patchFilePath := filepath.Join(sourceDir, "output.patch")
	if err := os.WriteFile(patchFilePath, []byte(cuestackBlueprint), 0644); err != nil {
		return fmt.Errorf("error writing to patch file: %v", err)
	}

	fmt.Printf("Cuestack blueprint generated successfully at: %s\n", patchFilePath)
	return nil
}

func SendToOpenAI(prompt string) (string, error) {
	// Simulate sending the prompt to OpenAI and receiving a response
	// Here we just create a dummy response as if it was processed by OpenAI
	dummyResponse := "Cuestack blueprint based on provided HCL content."

	// Normally, here you would have code to interact with OpenAI's API
	// For example:
	// client := openai.NewClient(apiKey)
	// response, err := client.Completions.Create(openai.CompletionRequest{
	// 	Model: "text-davinci-002",
	// 	Prompt: prompt,
	// 	MaxTokens: 1024,
	// })
	// if err != nil {
	// 	return "", fmt.Errorf("OpenAI API error: %v", err)
	// }
	// return response.Choices[0].Text, nil

	// Since we are mocking:
	return dummyResponse, nil
}

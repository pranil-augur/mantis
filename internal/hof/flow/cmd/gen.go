package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Gen(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: gen <target directory> <package name>")
	}
	targetDir, packageName := args[0], args[1]

	// Initialize the cue module
	cueModInitCmd := exec.Command("cue", "mod", "init", packageName)
	cueModInitCmd.Dir = targetDir
	if err := cueModInitCmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize cue module: %v", err)
	}

	// Define the structure of directories to be created
	directories := []string{
		"defs",
		"migrations",
		"policies",
		"schemas",
		"tests",
	}

	// Define files with initial content
	files := map[string]string{
		"defs/common.cue": `// Common definitions
`,
		"defs/data_sources.cue": `// Data source definitions
`,
		"defs/resources.cue": `
		package defs
		import "<module_name>/schemas"

eks: schemas.#ModuleVPC & {
    source:  "terraform-aws-modules/eks/aws"
    version: "~> 19.16"
}
`,
		"flows.tf.cue": `// Flow definitions
`,
		"migrations/migrations.tf.cue": `// Migration definitions
`,
		"migrations/v1.0.tf.cue": `// Version 1.0 migrations
`,
		"policies/costs.tf.cue": `// Cost policies
`,
		"policies/security.tf.cue": `// Security policies
`,
		"schemas/resource.cue": `
		package schemas
		#ModuleVPC: {
    source:  string
    version: string
}
`,
	}
	// Create directories
	for _, dir := range directories {
		dirPath := filepath.Join(targetDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dirPath, err)
		}
	}

	// Create files with initial content
	for file, content := range files {
		filePath := filepath.Join(targetDir, file)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %v", filePath, err)
		}
	}

	fmt.Println("Scaffolding generated successfully at:", targetDir)
	return nil
}

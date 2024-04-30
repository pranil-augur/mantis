// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configs

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/spf13/afero"
)

// Parser is the main interface to read configuration files and other related
// files from disk.
//
// It retains a cache of all files that are loaded so that they can be used
// to create source code snippets in diagnostics, etc.
type Parser struct {
	fs afero.Afero
	p  *hclparse.Parser
	c  *cue.Context

	// allowExperiments controls whether we will allow modules to opt in to
	// experimental language features. In main code this will be set only
	// for alpha releases and some development builds. Test code must decide
	// for itself whether to enable it so that tests can cover both the
	// allowed and not-allowed situations.
	allowExperiments bool
}

// NewParser creates and returns a new Parser that reads files from the given
// filesystem. If a nil filesystem is passed then the system's "real" filesystem
// will be used, via afero.OsFs.
func NewParser(fs afero.Fs) *Parser {
	if fs == nil {
		fs = afero.OsFs{}
	}

	return &Parser{
		fs: afero.Afero{Fs: fs},
		p:  hclparse.NewParser(),
		c:  cuecontext.New(),
	}
}

// LoadHCLFile is a low-level method that reads the file at the given path,
// parses it, and returns the hcl.Body representing its root. In many cases
// it is better to use one of the other Load*File methods on this type,
// which additionally decode the root body in some way and return a higher-level
// construct.
//
// If the file cannot be read at all -- e.g. because it does not exist -- then
// this method will return a nil body and error diagnostics. In this case
// callers may wish to ignore the provided error diagnostics and produce
// a more context-sensitive error instead.
//
// The file will be parsed using the HCL native syntax unless the filename
// ends with ".json", in which case the HCL JSON syntax will be used.
func (p *Parser) LoadHCLFile(path string) (hcl.Body, hcl.Diagnostics) {
	src, err := p.fs.ReadFile(path)

	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read file",
				Detail:   fmt.Sprintf("The file %q could not be read.", path),
			},
		}
	}

	var file *hcl.File
	var diags hcl.Diagnostics
	switch {
	case strings.HasSuffix(path, ".json"):
		file, diags = p.p.ParseJSON(src, path)
	case strings.HasSuffix(path, ".cue"):
		file, diags = p.LoadCUEDir(path)
	default:
		file, diags = p.p.ParseHCL(src, path)
	}

	// If the returned file or body is nil, then we'll return a non-nil empty
	// body so we'll meet our contract that nil means an error reading the file.
	if file == nil || file.Body == nil {
		return hcl.EmptyBody(), diags
	}

	return file.Body, diags
}

// LoadHCLString is a method that reads the content from a string instead of a file,
// parses it, and returns the hcl.Body representing its root based on the specified format type.
//
// The content will be parsed using the specified format type: HCL native syntax, HCL JSON syntax,
// or CUE syntax.
func (p *Parser) LoadHCLString(content string, formatType string) (hcl.Body, hcl.Diagnostics) {
	var file *hcl.File
	var diags hcl.Diagnostics

	switch formatType {
	case "json":
		file, diags = p.p.ParseJSON([]byte(content), "input.json")
	case "cue":
		fmt.Println("Parsing CUE content as string with content:", content)
		// Assuming LoadCUEDir can be adapted to handle string content directly for CUE format
		file, diags = p.LoadCUEString(content)
	default:
		file, diags = p.p.ParseHCL([]byte(content), "input.hcl")
	}

	// If the returned file or body is nil, then we'll return a non-nil empty
	// body so we'll meet our contract that nil means an error parsing the content.
	if file == nil || file.Body == nil {
		return hcl.EmptyBody(), diags
	}

	return file.Body, diags
}

// LoadCUEString is a helper method to parse CUE content from a string and convert it to HCL via JSON.
func (p *Parser) LoadCUEString(content string) (*hcl.File, hcl.Diagnostics) {
	c := cuecontext.New()
	instance := c.CompileString(content)
	if instance.Err() != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to compile CUE content",
				Detail:   fmt.Sprintf("Error compiling CUE content: %v", instance.Err()),
			},
		}
	}
	cueformValue := instance.LookupPath(cue.ParsePath("cueform"))

	// Convert the CUE instance to JSON
	jsonBytes, err := cueformValue.MarshalJSON()
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to convert CUE instance to JSON",
				Detail:   fmt.Sprintf("Error marshalling CUE instance to JSON: %v", err),
			},
		}
	}

	// Parse the JSON as HCL
	file, diags := p.p.ParseJSON(jsonBytes, "input.json")
	if diags.HasErrors() {
		return nil, diags
	}

	return file, nil
}

func (p *Parser) LoadCUEFileWrapper(path string) (hcl.Body, hcl.Diagnostics) {
	file, diags := p.LoadCUEDir(path)
	if file == nil {
		return hcl.EmptyBody(), diags
	}
	return file.Body, diags
}

func (p *Parser) LoadCUEDir(path string) (*hcl.File, hcl.Diagnostics) {
	var dirPath string
	if strings.HasSuffix(path, ".cue") {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return &hcl.File{}, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Failed to get absolute directory path",
					Detail:   fmt.Sprintf("Error obtaining absolute path for directory %q: %v", path, err),
				},
			}
		}
		dirPath = filepath.Dir(absPath)
	}

	c := cuecontext.New()
	// Assuming 'path' is the directory containing your CUE files
	cfg := &load.Config{
		Dir:         dirPath, // Set the directory to load all CUE files from
		Package:     "*",
		AllCUEFiles: true,
	}

	// Create instances from all CUE files in the directory, including imported modules
	buildInstances := load.Instances([]string{"."}, cfg) // Use absDirPath to indicate all files in the specified directory

	var mergedInstance cue.Value
	for _, inst := range buildInstances {
		if inst.Err != nil {
			log.Fatalf("Error loading CUE files: %v", inst.Err)
		}

		value := c.BuildInstance(inst)
		if value.Err() != nil {
			log.Fatalf("Error building CUE instance: %v", value.Err())
		}

		// Merge all instances into a single instance for unified handling
		if !mergedInstance.Exists() {
			mergedInstance = value
		} else {
			mergedInstance = mergedInstance.Unify(value)
			if mergedInstance.Err() != nil {
				log.Fatalf("Error merging CUE instances: %v", mergedInstance.Err())
			}
		}

	}

	if mergedInstance.Err() != nil {
		return &hcl.File{}, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to merge CUE instances",
				Detail:   fmt.Sprintf("Error merging CUE instances: %v", mergedInstance.Err()),
			},
		}
	}

	// Assuming 'blueprint' is a top-level field in your CUE structure
	cueformValue := mergedInstance.LookupPath(cue.ParsePath("cueform"))
	if !cueformValue.Exists() {
		log.Fatalf("cueform tag not found in CUE instance")
	}

	// Convert the merged CUE instance to JSON
	jsonBytes, err := cueformValue.MarshalJSON()
	if err != nil {
		return &hcl.File{}, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to convert merged CUE instance to JSON",
				Detail:   fmt.Sprintf("Error marshalling merged CUE instance to JSON: %v", err),
			},
		}
	}

	file, diags := p.p.ParseJSON(jsonBytes, path)
	if diags.HasErrors() {
		return &hcl.File{}, diags
	}
	// fmt.Printf("Parsed HCL file contents: %s\n", string(jsonBytes))

	// Return the Body of the parsed file, which is of type hcl.Body
	return file, nil
}

// Sources returns a map of the cached source buffers for all files that
// have been loaded through this parser, with source filenames (as requested
// when each file was opened) as the keys.
func (p *Parser) Sources() map[string][]byte {
	return p.p.Sources()
}

// ForceFileSource artificially adds source code to the cache of file sources,
// as if it had been loaded from the given filename.
//
// This should be used only in special situations where configuration is loaded
// some other way. Most callers should load configuration via methods of
// Parser, which will update the sources cache automatically.
func (p *Parser) ForceFileSource(filename string, src []byte) {
	// We'll make a synthetic hcl.File here just so we can reuse the
	// existing cache.
	p.p.AddFile(filename, &hcl.File{
		Body:  hcl.EmptyBody(),
		Bytes: src,
	})
}

// AllowLanguageExperiments specifies whether subsequent LoadConfigFile (and
// similar) calls will allow opting in to experimental language features.
//
// If this method is never called for a particular parser, the default behavior
// is to disallow language experiments.
//
// Main code should set this only for alpha or development builds. Test code
// is responsible for deciding for itself whether and how to call this
// method.
func (p *Parser) AllowLanguageExperiments(allowed bool) {
	p.allowExperiments = allowed
}

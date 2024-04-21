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

func (p *Parser) LoadCUEFileWrapper(path string) (hcl.Body, hcl.Diagnostics) {
	file, diags := p.LoadCUEDir(path)
	if file == nil {
		return hcl.EmptyBody(), diags
	}
	return file.Body, diags
}

func (p *Parser) LoadCUEDir(path string) (*hcl.File, hcl.Diagnostics) {
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
		path = filepath.Dir(absPath)
	}
	fmt.Println("Loading files from " + path)
	// Read the CUE file and create an instance
	/*
		// Create an instance from the CUE file
		buildInstances := load.Instances([]string{path}, nil)
		if len(buildInstances) == 0 || buildInstances[0].Err != nil {
			return &hcl.File{}, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Failed to load CUE file",
					Detail:   fmt.Sprintf("Error parsing CUE file: %v", buildInstances[0].Err),
				},
			}
		}*/

	c := cuecontext.New()
	// Assuming 'path' is the directory containing your CUE files
	cfg := &load.Config{
		Dir:         path, // Set the directory to load all CUE files from
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

		fmt.Printf("Loaded and built instance from: %s\n", inst.DisplayPath)
	}

	if len(buildInstances) == 0 {
		return &hcl.File{}, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to load CUE files",
				Detail:   "No CUE instances found in the directory.",
			},
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

	// Convert the merged CUE instance to JSON
	jsonBytes, err := mergedInstance.MarshalJSON()
	if err != nil {
		return &hcl.File{}, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to convert merged CUE instance to JSON",
				Detail:   fmt.Sprintf("Error marshalling merged CUE instance to JSON: %v", err),
			},
		}
	}

	// Parse the JSON bytes using the HCL parser to get an hcl.File
	file, diags := p.p.ParseJSON(jsonBytes, path)
	if diags.HasErrors() {
		return &hcl.File{}, diags
	}

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

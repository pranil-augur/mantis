package os

import (
	g_os "os"

	"cuelang.org/go/cue"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/lib/yagu"
)

type ReadGlobs struct{}

func NewReadGlobs(val cue.Value) (hofcontext.Runner, error) {
	return &ReadGlobs{}, nil
}

func (T *ReadGlobs) Run(ctx *hofcontext.Context) (interface{}, error) {

	val := ctx.Value

	patterns, err := extractGlobConfig(ctx, val)
	if err != nil {
		return nil, err
	}

	filepaths, err := yagu.FilesFromGlobs(patterns)
	if err != nil {
		return nil, err
	}

	data := map[string]string{}

	for _, file := range filepaths {
		bs, err := g_os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		data[file] = string(bs)
	}

	return map[string]interface{}{"files": data}, nil
}

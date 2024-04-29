package cmd

import (
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/runtime"
)

func findMaxLabelLen(R *runtime.Runtime, dflags flags.DatamodelPflagpole) int {
	max := 0
	for _, dm := range R.Datamodels {
		m := dm.FindMaxLabelLen(dflags)
		if m > max {
			max = m
		}
	}
	return max
}



package mod

import (
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/repos/cache"
)

func Clean(rflags flags.RootPflagpole) (error) {
	upgradeHofMods()

	return cache.CleanCache()
}

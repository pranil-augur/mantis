package help

import "github.com/opentofu/opentofu/internal/hof/lib/connector"

func New() connector.Connector {
	items := []interface{}{
		NewHelp(),
	}
	m := connector.New("help")
	m.Add(items)

	return m
}

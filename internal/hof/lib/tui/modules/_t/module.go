package {{ .name | lower }}

import "github.com/opentofu/opentofu/internal/hof/lib/connector"

func New() connector.Connector {
	items := []any{
		New{{ .name | title }}(),
	}
	m := connector.New("{{ .name | title }}")
	m.Add(items)

	return m
}


package shared

import "cuelang.org/go/cue"

type Match struct {
	Value    string
	CueValue cue.Value
	Path     string
	File     string
	Type     string
	Children []Match
}

type QueryResult struct {
	Matches map[string][]Match
}

type QueryConfig struct {
	From   string         `json:"from"`             // Data source path
	Select []string       `json:"select,omitempty"` // Fields to project
	Where  map[string]any `json:"where,omitempty"`  // Predicate conditions
}

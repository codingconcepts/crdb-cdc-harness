package random

import (
	"fmt"

	"github.com/codingconcepts/crdb-cdc-harness/pkg/models"
)

// GenerateArgValues creates random arguments for any command line --arg
// arguments provided.
func GenerateArgValues(args models.StringFlags) ([]any, error) {
	var out []any

	for _, arg := range args {
		v, ok := Generators[arg]
		if !ok {
			return nil, fmt.Errorf("missing generator %q", arg)
		}

		out = append(out, v())
	}

	return out, nil
}

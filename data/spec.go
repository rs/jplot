package data

import (
	"fmt"
	"strings"
)

// Spec specify a list of field for a single graph.
type Spec struct {
	Fields []Field
}

// Field describe a field in a graph.
type Field struct {
	ID        string
	Name      string
	IsCounter bool
	IsMarker  bool
}

// ParseSpec parses a graph specification. Each spec is a string with one or
// more JSON path separated by + with fields options prefixed with colon and
// separated by commas.
func ParseSpec(args []string) ([]Spec, error) {
	specs := make([]Spec, 0, len(args))
	for i, v := range args {
		spec := Spec{}
		for j, name := range strings.Split(v, "+") {
			var isCounter bool
			var isMarker bool
			if strings.HasPrefix(name, "marker:counter:") {
				// Backward compat.
				name = strings.Replace(name, "marker:counter:", "marker,counter:", 1)
			}
			if idx := strings.IndexByte(name, ':'); idx != -1 {
				options := strings.Split(name[:idx], ",")
				name = name[idx+1:]
				for _, o := range options {
					switch o {
					case "counter":
						isCounter = true
					case "marker":
						isMarker = true
					default:
						return nil, fmt.Errorf("invalid field option: %s", o)
					}
				}
			}
			spec.Fields = append(spec.Fields, Field{
				ID:        fmt.Sprintf("%d.%d.%s", i, j, name),
				Name:      name,
				IsCounter: isCounter,
				IsMarker:  isMarker,
			})
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

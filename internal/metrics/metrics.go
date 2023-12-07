package metrics

import "fmt"

type Metric struct {
	Kind  Kind
	Name  string
	Value string
}

func (m Metric) Key() string {
	return fmt.Sprintf("%s_%s", m.Kind, m.Name)
}

package metrics

import "fmt"

type Metric struct {
	Type  MetricType
	Name  string
	Value any
}

func (m Metric) Key() string {
	return fmt.Sprintf("%s_%s", m.Type, m.Name)
}

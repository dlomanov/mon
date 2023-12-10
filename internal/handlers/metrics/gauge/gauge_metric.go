package gauge

import (
	"fmt"
	"github.com/dlomanov/mon/internal/handlers/metrics"
	"strconv"
)

func init() {
	var _ metrics.Metric = (*Metric)(nil)
}

type Metric struct {
	Name  string
	Value float64
}

func (m Metric) StringValue() string {
	return strconv.FormatFloat(m.Value, 'f', -1, 64)
}

func (m Metric) Key() string {
	return fmt.Sprintf("%s_%s", metrics.MetricGauge, m.Name)
}

func NewMetric(name, valueString string) (metric Metric, err error) {
	value, err := strconv.ParseFloat(valueString, 64)
	metric = Metric{Name: name, Value: value}
	return
}

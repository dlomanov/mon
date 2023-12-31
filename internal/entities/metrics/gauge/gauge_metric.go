package gauge

import (
	"github.com/dlomanov/mon/internal/entities/metrics"
	"strconv"
)

func init() {
	var _ metrics.Metric = (*Metric)(nil)
}

type Metric struct {
	Name  string
	Value float64
}

func (m Metric) Deconstruct() (mtype, name, value string) {
	mtype = string(metrics.MetricGauge)
	name = m.Name
	value = m.StringValue()
	return
}

func (m Metric) StringValue() string {
	return strconv.FormatFloat(m.Value, 'f', -1, 64)
}

func (m Metric) Key() string {
	return metrics.MetricGauge.CreateKey(m.Name)
}

func NewMetric(name, valueString string) (metric Metric, err error) {
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return Metric{}, err
	}

	return Metric{Name: name, Value: value}, nil
}

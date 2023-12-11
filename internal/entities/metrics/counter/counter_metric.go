package counter

import (
	"github.com/dlomanov/mon/internal/entities/metrics"
	"strconv"
)

func init() {
	var _ metrics.Metric = (*Metric)(nil)
}

type Metric struct {
	Name  string
	Value int64
}

func (m Metric) Deconstruct() (mtype, name, value string) {
	mtype = string(metrics.MetricCounter)
	name = m.Name
	value = m.StringValue()
	return
}

func (m Metric) StringValue() string {
	return strconv.FormatInt(m.Value, 10)
}

func (m Metric) Key() string {
	return metrics.MetricCounter.CreateKey(m.Name)
}

func NewMetric(name, valueString string) (metric Metric, err error) {
	return newMetric(name, valueString)
}

func (m Metric) With(valueString string) (metric Metric, err error) {
	return newMetric(m.Name, valueString)
}

func newMetric(name, valueString string) (metric Metric, err error) {
	value, err := strconv.ParseInt(valueString, 10, 64)
	metric = Metric{Name: name, Value: value}
	return
}

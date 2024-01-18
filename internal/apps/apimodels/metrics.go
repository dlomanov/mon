package apimodels

type Metric struct {
	MetricKey
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type MetricKey struct {
	Name string `json:"id"`
	Type string `json:"type"`
}

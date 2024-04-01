// Package apimodels provides API models for the application.
package apimodels

// Metric represents a metric with a key and optional delta or value.
// It is used to store and retrieve metrics in the application.
type Metric struct {
	MetricKey
	Delta *int64   `json:"delta,omitempty"` // Delta is the change in value for a counter metric.
	Value *float64 `json:"value,omitempty"` // Value is the current value for a gauge metric.
}

// MetricKey is a unique identifier for a metric, consisting of a name and type.
type MetricKey struct {
	Name string `json:"id"`   // Name is the unique name of the metric.
	Type string `json:"type"` // Type is the type of the metric (e.g., "counter", "gauge").
}

package metrics

type Metric interface {
	Key() string
	StringValue() string
}

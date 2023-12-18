package metrics

type Metric interface {
	Key() string
	StringValue() string
	Deconstruct() (mtype, name, value string)
}

package metrics

import "strings"

type Kind string

const (
	KindGauge   Kind = "gauge"
	KindCounter Kind = "counter"
)

func (k Kind) IsValid() bool {
	return k == KindGauge || k == KindCounter
}

func (k Kind) String() string {
	return string(k)
}

func ParseKind(str string) (value Kind, ok bool) {
	switch lower := strings.ToLower(str); lower {
	case string(KindGauge):
		return KindGauge, true
	case string(KindCounter):
		return KindCounter, true
	default:
		return "", false
	}
}

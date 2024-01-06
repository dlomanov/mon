package apperrors

type AppErrorType uint64

func (t AppErrorType) New(args ...any) AppError {
	return AppError{
		Type:       t,
		Message:    templateMap[t],
		FormatArgs: args,
	}
}

func (t AppErrorType) Newf(message string, args ...any) AppError {
	return AppError{
		Type:       t,
		Message:    message,
		FormatArgs: args,
	}
}

const (
	// ErrUnsupportedMetricType "unsupported metric type %s"
	ErrUnsupportedMetricType AppErrorType = iota + 1

	// ErrInvalidMetricPath "invalid metric path %s"
	ErrInvalidMetricPath

	// ErrInvalidMetricType "invalid metric type %s"
	ErrInvalidMetricType

	// ErrInvalidMetricName "invalid metric name %s"
	ErrInvalidMetricName

	// ErrInvalidMetricValue "invalid metric value %s"
	ErrInvalidMetricValue
)

var templateMap = map[AppErrorType]string{
	ErrUnsupportedMetricType: "unsupported metric type %s",
	ErrInvalidMetricPath:     "invalid metric path %s",
	ErrInvalidMetricType:     "invalid metric type %s",
	ErrInvalidMetricName:     "invalid metric name %s",
	ErrInvalidMetricValue:    "invalid metric value %s",
}

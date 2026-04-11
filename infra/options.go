package infra

// MergeOptions applies a slice of functional options to a zero-value T and
// returns the merged result. This eliminates the repeated merge-loop pattern
// across LoggerOption, TracingOption, ValidateOption, etc.
func MergeOptions[T any, O ~func(*T)](opts ...O) T {
	var result T
	for _, opt := range opts {
		opt(&result)
	}
	return result
}

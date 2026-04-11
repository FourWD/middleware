package infra

import (
	"errors"
	"slices"
	"strings"
)

type ValidationCollector struct {
	errors map[string][]error
}

type ValidationOptions struct {
	errorTransforms []func(error) (error, bool)
}

type ValidateOption func(*ValidationOptions)

type ValidationError struct {
	errors []ValidationErrorEntry
}

type ValidationErrorEntry struct {
	property string
	code     string
	reason   string
}

type CodedError struct {
	err  error
	code string
}

func NewValidationCollector() *ValidationCollector {
	return &ValidationCollector{
		errors: make(map[string][]error),
	}
}

func (v *ValidationCollector) AddError(property string, err error) {
	v.errors[property] = append(v.errors[property], err)
}

func (v *ValidationCollector) Validate(options ...ValidateOption) *ValidationError {
	if len(v.errors) == 0 {
		return nil
	}

	opts := MergeOptions[ValidationOptions](options...)
	result := &ValidationError{}

	for property, errs := range v.errors {
		for _, current := range errs {
			var codedError *CodedError
			if errors.As(current, &codedError) {
				result.errors = append(result.errors, ValidationErrorEntry{
					property: property,
					code:     codedError.Code(),
					reason:   codedError.Error(),
				})
				continue
			}

			transformed := current
			for _, transform := range opts.errorTransforms {
				next, done := transform(transformed)
				transformed = next
				if done {
					break
				}
			}

			if errors.As(transformed, &codedError) {
				result.errors = append(result.errors, ValidationErrorEntry{
					property: property,
					code:     codedError.Code(),
					reason:   codedError.Error(),
				})
				continue
			}

			result.errors = append(result.errors, ValidationErrorEntry{
				property: property,
				code:     "UNHANDLED",
				reason:   current.Error(),
			})
		}
	}

	return result
}

func NewValidationError(errors []ValidationErrorEntry) ValidationError {
	return ValidationError{errors: errors}
}

func (e ValidationError) Errors() []ValidationErrorEntry {
	return e.errors
}

func (e ValidationError) Error() string {
	return "validation error"
}

func NewValidationErrorEntry(property, code, reason string) ValidationErrorEntry {
	return ValidationErrorEntry{
		property: property,
		code:     code,
		reason:   reason,
	}
}

func (e ValidationErrorEntry) Property() string { return e.property }
func (e ValidationErrorEntry) Code() string     { return e.code }
func (e ValidationErrorEntry) Reason() string   { return e.reason }

func NewCodedError(err error, code string) *CodedError {
	return &CodedError{err: err, code: code}
}

func (e *CodedError) Code() string {
	return e.code
}

func (e *CodedError) Error() string {
	return e.err.Error()
}

func (e *CodedError) Unwrap() error {
	return e.err
}

func WithGlobalErrorCode(target error, code string) ValidateOption {
	return func(opts *ValidationOptions) {
		opts.errorTransforms = append(opts.errorTransforms, func(err error) (error, bool) {
			if errors.Is(err, target) {
				return NewCodedError(err, code), true
			}
			return err, false
		})
	}
}

func WithGlobalErrorCodeFunc(fn func(error) bool, code string) ValidateOption {
	return func(opts *ValidationOptions) {
		opts.errorTransforms = append(opts.errorTransforms, func(err error) (error, bool) {
			if fn(err) {
				return NewCodedError(err, code), true
			}
			return err, false
		})
	}
}

func WithGlobalErrorCodeAs[T error](code string) ValidateOption {
	return func(opts *ValidationOptions) {
		opts.errorTransforms = append(opts.errorTransforms, func(err error) (error, bool) {
			var target T
			if errors.As(err, &target) {
				return NewCodedError(err, code), true
			}
			return err, false
		})
	}
}

func WithGlobalCodeContains(fragment, code string) ValidateOption {
	return func(opts *ValidationOptions) {
		opts.errorTransforms = append(opts.errorTransforms, func(err error) (error, bool) {
			if strings.Contains(err.Error(), fragment) {
				return NewCodedError(err, code), true
			}
			return err, false
		})
	}
}

func MakeValidationErrorDeterministic(err *ValidationError) *ValidationError {
	cloned := append(([]ValidationErrorEntry)(nil), err.errors...)
	slices.SortFunc(cloned, func(a, b ValidationErrorEntry) int {
		if v := strings.Compare(a.property, b.property); v != 0 {
			return v
		}
		if v := strings.Compare(a.code, b.code); v != 0 {
			return v
		}
		return strings.Compare(a.reason, b.reason)
	})
	return &ValidationError{errors: cloned}
}

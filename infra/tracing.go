package infra

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type IgnoreErrorStruct struct {
	err error
}

func (e IgnoreErrorStruct) Is(target error) bool {
	return errors.Is(e.err, target)
}

func (e IgnoreErrorStruct) As(target any) bool {
	return errors.As(e.err, target)
}

func (e IgnoreErrorStruct) Error() string {
	return e.err.Error()
}

func IgnoreError(err error) error {
	if err == nil {
		return nil
	}

	return IgnoreErrorStruct{err: err}
}

type InvokeDurationCallback func(elapsed time.Duration, err error)

type TracingOptions struct {
	invokeDurationCallback InvokeDurationCallback
	mustSample             bool
}

type TracingOption func(*TracingOptions)

func WithInvokeDurationCallback(cb InvokeDurationCallback) TracingOption {
	return func(opts *TracingOptions) {
		opts.invokeDurationCallback = cb
	}
}

func MustSample() TracingOption {
	return func(opts *TracingOptions) {
		opts.mustSample = true
	}
}

func Trace(
	tracer trace.Tracer,
	ctx context.Context,
	name string,
	fn func(ctx context.Context, span trace.Span) error,
	options ...TracingOption,
) error {
	_, err := TraceResult(tracer, ctx, name, func(ctx context.Context, span trace.Span) (struct{}, error) {
		return struct{}{}, fn(ctx, span)
	}, options...)
	return err
}

func TraceResult[TResult any](
	tracer trace.Tracer,
	ctx context.Context,
	name string,
	fn func(ctx context.Context, span trace.Span) (TResult, error),
	options ...TracingOption,
) (TResult, error) {
	var (
		result TResult
		err    error
	)

	opts := MergeOptions[TracingOptions](options...)

	if opts.invokeDurationCallback != nil {
		start := time.Now()
		defer func() {
			opts.invokeDurationCallback(time.Since(start), err)
		}()
	}

	ctx, span := tracer.Start(ctx, name)
	defer span.End()

	if opts.mustSample {
		span.SetAttributes(attribute.Bool("must_sample", true))
	}

	result, err = fn(ctx, span)
	if err != nil {
		if ignored, ok := err.(IgnoreErrorStruct); ok {
			err = ignored.err
		} else {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}
	}

	return result, err
}

package kit

// Result is the shared success/error object standard for the boilerplate.
// Keep normal Go `(T, error)` returns as the default and use Result when
// success/error must travel together as a first-class value.

type Result[T any] struct {
	ok    bool
	value T
	err   error
}

func Ok[T any](value T) Result[T] {
	return Result[T]{ok: true, value: value}
}

func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

func (r Result[T]) IsOk() bool {
	return r.ok
}

func (r Result[T]) IsErr() bool {
	return !r.ok
}

func (r Result[T]) TryUnwrap() (T, error) {
	return r.value, r.err
}

func (r Result[T]) TryGetValue() (T, bool) {
	return r.value, r.ok
}

func (r Result[T]) TryGetError() (error, bool) {
	return r.err, !r.ok
}

func (r Result[T]) UnsafeUnwrapOk() T {
	if r.ok {
		return r.value
	}
	panic("unwrap ok called on err result")
}

func (r Result[T]) UnsafeUnwrapErr() error {
	if !r.ok {
		return r.err
	}
	panic("unwrap err called on ok result")
}

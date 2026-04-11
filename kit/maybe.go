package kit

import "errors"

// Maybe is the shared optional-value standard for the boilerplate.
// Use it when explicit optionality is clearer than zero values or nil.

var ErrUnwrapNoneValue = errors.New("unwrap none value")

type Maybe[T any] struct {
	set   bool
	value T
}

func Some[T any](value T) Maybe[T] {
	return Maybe[T]{set: true, value: value}
}

func None[T any]() Maybe[T] {
	return Maybe[T]{}
}

func (m Maybe[T]) IsSome() bool {
	return m.set
}

func (m Maybe[T]) IsNone() bool {
	return !m.set
}

func (m Maybe[T]) TryGetValue() (T, bool) {
	if m.set {
		return m.value, true
	}
	var zero T
	return zero, false
}

func (m Maybe[T]) UnsafeUnwrap() T {
	if !m.set {
		panic(ErrUnwrapNoneValue)
	}
	return m.value
}

func (m Maybe[T]) UnwrapOr(value T) T {
	if m.set {
		return m.value
	}
	return value
}

func (m Maybe[T]) UnwrapOrZero() T {
	if m.set {
		return m.value
	}
	var zero T
	return zero
}

func MaybeToPointer[T any](m Maybe[T]) *T {
	if m.set {
		return &m.value
	}
	return nil
}

func MaybeFromPointer[T any](value *T) Maybe[T] {
	if value == nil {
		return None[T]()
	}
	return Some(*value)
}

func MapMaybe[T any, U any](m Maybe[T], fn func(T) U) Maybe[U] {
	if m.set {
		return Some(fn(m.value))
	}
	return None[U]()
}

func TryMapMaybe[T any, U any](m Maybe[T], fn func(T) (U, error)) (Maybe[U], error) {
	if !m.set {
		return None[U](), nil
	}

	value, err := fn(m.value)
	if err != nil {
		return None[U](), err
	}

	return Some(value), nil
}

func FlatMapMaybe[T any, U any](m Maybe[T], fn func(T) Maybe[U]) Maybe[U] {
	if m.set {
		return fn(m.value)
	}
	return None[U]()
}

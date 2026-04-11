package kit

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](elements ...T) Set[T] {
	set := make(Set[T], len(elements))
	for _, element := range elements {
		set[element] = struct{}{}
	}
	return set
}

func (s Set[T]) Add(element T) {
	s[element] = struct{}{}
}

func (s Set[T]) Remove(element T) {
	delete(s, element)
}

func (s Set[T]) Has(element T) bool {
	_, ok := s[element]
	return ok
}

func (s Set[T]) DoesNotHave(element T) bool {
	return !s.Has(element)
}

func (s Set[T]) Size() int {
	return len(s)
}

func (s Set[T]) IsEmpty() bool {
	return s.Size() == 0
}

func (s Set[T]) IsNotEmpty() bool {
	return s.Size() > 0
}

func (s Set[T]) Elements() []T {
	result := make([]T, 0, len(s))
	for element := range s {
		result = append(result, element)
	}
	return result
}

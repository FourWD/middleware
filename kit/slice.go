package kit

func SliceContains[T comparable](slice []T, value T) bool {
	for _, current := range slice {
		if current == value {
			return true
		}
	}
	return false
}

func SliceFilter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, current := range slice {
		if predicate(current) {
			result = append(result, current)
		}
	}
	return result
}

func SliceMap[T any, U any](slice []T, fn func(T) U) []U {
	return SliceMapWithIndex(slice, func(_ int, value T) U {
		return fn(value)
	})
}

func SliceMapWithIndex[T any, U any](slice []T, fn func(int, T) U) []U {
	result := make([]U, len(slice))
	for index, current := range slice {
		result[index] = fn(index, current)
	}
	return result
}

func SliceTryMap[T any, U any](slice []T, fn func(T) (U, error)) ([]U, error) {
	return SliceTryMapWithIndex(slice, func(_ int, value T) (U, error) {
		return fn(value)
	})
}

func SliceTryMapWithIndex[T any, U any](slice []T, fn func(int, T) (U, error)) ([]U, error) {
	result := make([]U, len(slice))
	for index, current := range slice {
		value, err := fn(index, current)
		if err != nil {
			return nil, err
		}
		result[index] = value
	}
	return result, nil
}

func SliceChunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}

	result := make([][]T, 0)
	for start := 0; start < len(slice); start += size {
		end := start + size
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[start:end])
	}
	return result
}

func SliceGroupBy[K comparable, T any](slice []T, keyFunc func(T) K) map[K][]T {
	grouped := make(map[K][]T)
	for _, item := range slice {
		key := keyFunc(item)
		grouped[key] = append(grouped[key], item)
	}
	return grouped
}

func SliceKeyBy[T any, V any](slice []T, keyFunc func(T) (string, V)) map[string]V {
	result := make(map[string]V)
	for _, item := range slice {
		key, value := keyFunc(item)
		result[key] = value
	}
	return result
}

func SliceTryKeyBy[T any, V any](slice []T, keyFunc func(T) (string, V, error)) (map[string]V, error) {
	result := make(map[string]V)
	for _, item := range slice {
		key, value, err := keyFunc(item)
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

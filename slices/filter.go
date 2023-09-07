package slices

func Filter[T any](s []T, fn func(T) bool) []T {
	result := []T{}
	for _, elem := range s {
		if fn(elem) {
			result = append(result, elem)
		}
	}
	return result
}

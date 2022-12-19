package helpers

// isNil checks if the given value is nil.
func isNil(value interface{}) bool {
	return value == nil
}

// isEmpty checks if the given value is an empty string or a slice/array with no elements.
func isEmpty(value interface{}) bool {
	switch value := value.(type) {
	case string:
		return value == ""
	case []interface{}, []string:
		return len(value.([]interface{})) == 0
	default:
		return false
	}
}

// isUndefined checks if the given value is undefined.
func isUndefined[T comparable](value T) bool {
	return value == *new(T)
	// switch value.(type) {
	// case nil:
	// 	return true
	// default:
	// 	return false
	// }
}
func Ternary[T comparable](value1, value2 T) T {
	if isNil(value1) || isEmpty(value1) || isUndefined(value1) {
		return value2
	}
	return value1
}

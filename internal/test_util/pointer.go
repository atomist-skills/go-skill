package test_util

// Pointer is useful for making pointers of literals in test cases
// e.g. Pointer(3) or Pointer("string")
func Pointer[T any](some T) *T {
	return &some
}

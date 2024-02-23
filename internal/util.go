package internal

func ContainsBy[K interface{}](slice []K, test func(K) bool) bool {
	for _, v := range slice {
		if test(v) {
			return true
		}
	}
	return false
}

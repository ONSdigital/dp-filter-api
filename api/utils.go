package api

// createArray creates an array of keys from the provided map
func createArray(m map[string]struct{}) (a []string) {
	for k := range m {
		a = append(a, k)
	}
	return a
}

// createMap creates a map whose keys are the unique values of the provided array(s).
// values are empty structs for memory efficiency reasons (no storage used)
func createMap(a ...[]string) (m map[string]struct{}) {
	m = make(map[string]struct{})
	for _, aa := range a {
		for _, val := range aa {
			m[val] = struct{}{}
		}
	}
	return m
}

// return the lowest int value
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

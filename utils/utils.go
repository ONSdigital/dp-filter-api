package utils

// CreateArray creates an array of keys from the provided map
func CreateArray(m map[string]struct{}) []string {
	array := []string{}
	for k := range m {
		array = append(array, k)
	}
	return array
}

// CreateMap creates a map whose keys are the unique values of the provided array(s).
// values are empty structs for memory efficiency reasons (no storage used)
func CreateMap(a ...[]string) (m map[string]struct{}) {
	m = make(map[string]struct{})
	for _, aa := range a {
		for _, val := range aa {
			m[val] = struct{}{}
		}
	}
	return m
}

// Min returns the lowest int value
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

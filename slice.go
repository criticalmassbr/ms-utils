package utils

func HaveSameElements[T comparable](original []T, expected []T) bool {
	if len(original) != len(expected) {
		return false
	}

	for _, e := range original {
		if !Contains(expected, e) {
			return false
		}
	}

	return true
}

func Contains[T comparable](original []T, expected T) bool {
	for _, e := range original {
		if e == expected {
			return true
		}
	}
	return false
}

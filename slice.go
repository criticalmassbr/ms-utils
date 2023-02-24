package utils

import (
	"sort"

	"github.com/google/uuid"
	"golang.org/x/exp/constraints"
)

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

func RemoveDuplicatesFromOrderedSlice[T constraints.Ordered](a []T) []T {
	if len(a) == 0 {
		return a
	}

	j := 0
	for i := 1; i < len(a); i++ {
		if a[j] != a[i] {
			j++
			a[j] = a[i]
		}
	}
	return a[:j+1]
}

func SortAndRemoveDuplicates[T constraints.Ordered](a []T) []T {
	if len(a) == 0 {
		return a
	}

	sort.Slice(a, func(i, j int) bool {
		return a[i] < a[j]
	})
	return RemoveDuplicatesFromOrderedSlice(a)
}

func SortAndRemoveDuplicateUUIDs(a []uuid.UUID) []uuid.UUID {
	if len(a) == 0 {
		return a
	}

	sort.Slice(a, func(i, j int) bool {
		return a[i].String() < a[j].String()
	})
	j := 0
	for i := 1; i < len(a); i++ {
		if a[j].String() != a[i].String() {
			j++
			a[j] = a[i]
		}
	}
	return a[:j+1]
}

func SortAndRemoveDuplicateConditions(a []Condition) []Condition {
	if len(a) == 0 {
		return a
	}

	sort.Slice(a, func(i, j int) bool {
		if a[i].FieldName == a[j].FieldName {
			return a[i].Operator < a[j].Operator
		}
		return a[i].FieldName < a[j].FieldName
	})
	j := 0
	for i := 1; i < len(a); i++ {
		if a[j].FieldName != a[i].FieldName || a[j].Operator != a[i].Operator {
			j++
			a[j] = a[i]
		}
	}
	return a[:j+1]
}

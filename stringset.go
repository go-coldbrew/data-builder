package databuilder

import "sort"

// stringSet is an internal replacement for k8s.io/apimachinery/pkg/util/sets.String.
// It implements only the subset of methods used by data-builder.
type stringSet map[string]struct{}

// newStringSet creates a stringSet from a list of values.
func newStringSet(items ...string) stringSet {
	s := make(stringSet, len(items))
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

// Has returns true if the set contains the given item.
func (s stringSet) Has(item string) bool {
	_, found := s[item]
	return found
}

// Insert adds items to the set.
func (s stringSet) Insert(items ...string) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

// IsSuperset returns true if s contains all items in other.
func (s stringSet) IsSuperset(other stringSet) bool {
	for item := range other {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// Len returns the number of items in the set.
func (s stringSet) Len() int {
	return len(s)
}

// List returns sorted list of items in the set.
func (s stringSet) List() []string {
	res := make([]string, 0, len(s))
	for item := range s {
		res = append(res, item)
	}
	sort.Strings(res)
	return res
}

// String returns a human-readable representation of the set, e.g. "[a b c]".
func (s stringSet) String() string {
	items := s.List()
	result := "["
	for i, item := range items {
		if i > 0 {
			result += " "
		}
		result += item
	}
	result += "]"
	return result
}

// Difference returns a new set with items in s but not in other.
func (s stringSet) Difference(other stringSet) stringSet {
	result := newStringSet()
	for item := range s {
		if !other.Has(item) {
			result.Insert(item)
		}
	}
	return result
}

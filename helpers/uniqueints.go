package helpers

import "sort"

type (
	UniqueInts map[int]struct{}
)

//AddValue to unique list
func (u *UniqueInts) AddValue(v int) {
	if u == nil || *u == nil {
		*u = make(map[int]struct{})
	}
	var emptyStruct struct{}
	(*u)[v] = emptyStruct
}

//Array of unique integers
func (u *UniqueInts) Array() []int {
	if u == nil || *u == nil {
		return []int{}
	}

	a := make([]int, len(*u))
	a = a[:0]

	for k := range *u {
		a = append(a, k)
	}
	// sort the keys so seeded randomization always gets the same order
	sort.Ints(a)
	return a
}

//HasValue test if collection includes value
func (u *UniqueInts) HasValue(v int) bool {
	if u == nil || *u == nil {
		return false
	}
	_, exist := (*u)[v]
	return exist
}

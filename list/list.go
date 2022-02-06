package list

func Filter(in []int, predicate func(int) bool) []int {

	out := []int{}

	for _, n := range in {
		if predicate(n) {
			out = append(out, n)
		}
	}

	return out
}

func Includes(in []int, n int) bool {
	for _, c := range in {
		if c == n {
			return true
		}
	}
	return false
}

//go 1.18 generics

// func Map[T int | int[]}(in []T, mapper func(int) T) []T {

// 	out := make([]int, len(in))

// 	for i, n := range in {
// 		out[i] = mapper(n)
// 	}

// 	return out
// }

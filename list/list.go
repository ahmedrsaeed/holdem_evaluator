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

func ValuesAtIndexes(in []int, indexes []int) []int {
	values := make([]int, len(indexes))

	for i, n := range indexes {
		values[i] = in[n]
	}

	return values
}

func CopyValuesAtIndexes(dst []int, src []int, indexes []int) {

	for dstI := 0; dstI < len(indexes); dstI++ {
		srcI := indexes[dstI]
		dst[dstI] = src[srcI]
	}
}

func Clone(in []int) []int {
	out := make([]int, len(in))
	copy(out, in)
	return out
}

//go 1.18 generics

// func Map[T int | int[]}(in []T, mapper func(int) T) []T {

// 	out := make([]int, len(in))

// 	for i, n := range in {
// 		out[i] = mapper(n)
// 	}

// 	return out
// }

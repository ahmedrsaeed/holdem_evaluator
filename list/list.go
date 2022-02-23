package list

func Filter(in []uint8, predicate func(uint8) bool) []uint8 {

	out := []uint8{}

	for _, n := range in {
		if predicate(n) {
			out = append(out, n)
		}
	}

	return out
}

func Includes(in []uint8, n uint8) bool {
	for _, c := range in {
		if c == n {
			return true
		}
	}
	return false
}

func CopyValuesAtIndexes(dst []uint8, src []uint8, indexes []uint8) {

	for dstI := 0; dstI < len(indexes); dstI++ {
		dst[dstI] = src[indexes[dstI]]
	}
}

func Clone(in []uint8) []uint8 {
	out := make([]uint8, len(in))
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

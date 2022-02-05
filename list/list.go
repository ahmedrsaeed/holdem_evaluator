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

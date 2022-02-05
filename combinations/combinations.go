package combinations

import (
	"errors"
	"holdem/list"
)

type Combinations struct {
	store map[string][][]int
}

func New() Combinations {
	h := Combinations{}
	h.intialize()
	return h
}

func (c *Combinations) intialize() {

	expected := [][]int{
		{50, 2},
		{48, 5},
		{45, 1},
		{44, 1},
	}

	c.store = map[string][][]int{}
	for _, e := range expected {

		key, val := generate(e[0], e[1])
		c.store[key] = val
	}
}

func generate(n int, r int) (string, [][]int) {
	all := [][]int{}

	var helper func([]int, []int)

	helper = func(available []int, current []int) {

		if len(current) == r {
			all = append(all, current)
			return
		}

		lastCardIndex := len(current) - 1

		for _, c := range available {

			if lastCardIndex > -1 && current[lastCardIndex] > c {
				continue
			}

			helper(list.Filter(available, func(a int) bool {
				return a != c
			}), append(current, c))
		}
	}

	available := make([]int, n)
	for i := range available {
		available[i] = i
	}

	helper(available, make([]int, 0))

	return key(n, r), all
}

func (c *Combinations) Generate(n int, r int) ([][]int, error) {

	res, err := c.store[key(n, r)]

	if err {
		return nil, errors.New("unable to compute " + key(n, r))
	}

	return res, nil
}

func key(n int, r int) string {
	return `${n}c${r}`
}

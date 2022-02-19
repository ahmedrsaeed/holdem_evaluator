package combinations

import (
	"errors"
	"fmt"
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
		{47, 2},
		{46, 2},
		{45, 2},
		{44, 1},
		{43, 0},
	}

	c.store = map[string][][]int{}
	for _, e := range expected {

		key, val := generate(e[0], e[1])
		c.store[key] = val

	}
}

func generate(n int, r int) (string, [][]int) {
	all := [][]int{}

	if (r == 0) || r > n {
		all = append(all, []int{})
		return key(n, r), all
	}

	limits := make([]int, r)
	current := make([]int, r)
	for i := range limits {
		limits[i] = i + n - r
		current[i] = i
	}
	mod_index_end := r - 1
	mod_index := mod_index_end

Outer:
	for {

		for mod_index < mod_index_end {
			mod_index += 1
			current[mod_index] = current[mod_index-1] + 1
		}

		curr := make([]int, r)
		copy(curr, current)

		//fmt.Println(curr)
		all = append(all, curr)

		for {

			if current[mod_index] < limits[mod_index] {
				current[mod_index] += 1
				break
			} else if mod_index > 0 {
				mod_index -= 1
			} else {
				break Outer
			}
		}
	}

	//fmt.Println(n, r, len(all))
	return key(n, r), all
}

func (c *Combinations) Get(n int, r int) ([][]int, error) {

	res, ok := c.store[key(n, r)]

	if !ok {
		return nil, errors.New("unable to compute " + key(n, r))
	}

	return res, nil
}

func key(n int, r int) string {
	return fmt.Sprintf("%dc%d", n, r)
}

package combinations

import (
	"errors"
	"fmt"
	"holdem/list"
	"time"
)

type Combinations struct {
	store map[string][]Combination
}

type Combination struct {
	Selected []int
	Other    []int
}

func New() Combinations {
	h := Combinations{}
	h.intialize()
	return h
}

func (c *Combinations) intialize() {

	villainCombinations := [][]int{
		{50, 5},
		{47, 2},
		{46, 1},
		{45, 2},
		{43, 2},
		{41, 2},
		{39, 2},
		{37, 2},
		{35, 2},
		{33, 2},
		{31, 2},
		{29, 2},
		{27, 2},
	}

	c.store = map[string][]Combination{}
	for _, e := range villainCombinations {

		t := time.Now()
		key, combos := generate(e[0], e[1])
		fmt.Printf("%v t:%f\n", e, time.Since(t).Minutes())
		c.store[key] = combos
	}
}

func generate(n int, r int) (string, []Combination) {
	combinations := []Combination{}

	allIndexes := make([]int, n)
	for i := range allIndexes {
		allIndexes[i] = i
	}

	if (r == 0) || r > n {
		combinations = append(combinations, Combination{Selected: []int{}, Other: allIndexes})
		return key(n, r), combinations
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

		currSelected := list.Clone(current)

		other := list.Filter(allIndexes, func(i int) bool {
			return !list.Includes(currSelected, i)
		})

		//fmt.Printf("%v\n", current)
		combinations = append(combinations, Combination{Selected: currSelected, Other: other})

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
	return key(n, r), combinations
}

func key(n int, r int) string {
	return fmt.Sprintf("%dc%d", n, r)
}

func (c *Combinations) Get(n int, r int) ([]Combination, error) {

	res, ok := c.store[key(n, r)]

	if !ok {
		return nil, errors.New("unable to compute " + key(n, r))
	}

	return res, nil
}

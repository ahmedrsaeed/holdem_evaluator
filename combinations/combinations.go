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
	Selected []uint8
	Other    []uint8
}

func New() Combinations {
	h := Combinations{}
	h.intialize()
	return h
}

func (c *Combinations) intialize() {

	villainCombinations := [][]uint8{
		{52, 2},
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
		fmt.Printf("%v t:%f l:%d\n", e, time.Since(t).Minutes(), len(combos))
		if len(combos) > 1<<31-1 {
			panic("can't use Rand.Int31 for sampling")
		}
		c.store[key] = combos
	}
}

func generate(n uint8, r uint8) (string, []Combination) {
	combinations := []Combination{}

	allIndexes := make([]uint8, n)
	for i := range allIndexes {
		allIndexes[i] = uint8(i)
	}

	if (r == 0) || r > n {
		combinations = append(combinations, Combination{Selected: []uint8{}, Other: allIndexes})
		return key(n, r), combinations
	}

	limits := make([]uint8, r)
	current := make([]uint8, r)
	for i := range limits {
		limits[i] = uint8(i) + n - r
		current[i] = uint8(i)
	}
	var mod_index_end int = int(r) - 1
	var mod_index int = mod_index_end //int to allow for negative mod_index

Outer:
	for {

		for mod_index < mod_index_end {
			mod_index += 1
			current[mod_index] = current[mod_index-1] + 1
		}

		currSelected := list.Clone(current)

		other := list.Filter(allIndexes, func(i uint8) bool {
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

func key(n uint8, r uint8) string {
	return fmt.Sprintf("%dc%d", n, r)
}

func (c *Combinations) Get(n uint8, r uint8) ([]Combination, error) {

	res, ok := c.store[key(n, r)]

	if !ok {
		return nil, errors.New("unable to compute " + key(n, r))
	}

	return res, nil
}

// func (c *Combinations) GetAllPossiblePairs(available []int) ([][]int, map[int]map[int]int, error) {
// 	combos, err := c.Get(len(available), 2)

// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	pairs := make([][]int, len(combos))

// 	for i, combo := range combos {

// 		pairs[i] = list.ValuesAtIndexes(available, combo.Selected)
// 	}

// 	pairsIndexMap := make(map[int]map[int]int)

// 	for _, a := range available {
// 		pairsIndexMap[a] = make(map[int]int)
// 	}

// 	for i, p := range pairs {
// 		pairsIndexMap[p[0]][p[1]] = i
// 	}

// 	return pairs, pairsIndexMap, nil
// }

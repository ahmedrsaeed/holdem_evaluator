package combinationssampler

import (
	"holdem/combinations"
	"math/rand"
	"time"
)

type CombinationsSamplerCreator struct {
	//sampleIndexes map[int]struct{}
	rGen *rand.Rand
}

func NewCreator() CombinationsSamplerCreator {
	return CombinationsSamplerCreator{
		//sampleIndexes: make(map[int]struct{}),
		rGen: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

var exists = struct{}{}

func (creator *CombinationsSamplerCreator) Create(combos []combinations.Combination, desired int) (func(func(combinations.Combination)), int) {

	combinationsLength := len(combos)

	if combinationsLength <= desired {

		return func(action func(combinations.Combination)) {
			for _, combo := range combos {
				action(combo)
			}
		}, combinationsLength
	}

	return func(action func(combinations.Combination)) {

		// if len(creator.sampleIndexes) > 0 {
		// 	panic("combinations sampler creator resused without cleanup (concurrent or recursive use)")
		// }

		sampleIndexes := make(map[int]struct{})

		for len(sampleIndexes) < desired {

			randomIndex := creator.rGen.Intn(combinationsLength)

			if _, ok := sampleIndexes[randomIndex]; ok {
				continue
			}
			sampleIndexes[randomIndex] = exists
			action(combos[randomIndex])
		}

		// for selectedIndex := range sampleIndexes {
		// 	action(combos[selectedIndex])
		// 	//delete(creator.sampleIndexes, selectedIndex)
		// }
	}, desired
}

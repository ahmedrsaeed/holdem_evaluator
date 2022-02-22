package slicesampler

import (
	"math/rand"
	"time"
)

type Sampler struct {
	sliceLength        int
	desired            int
	sampleIndexes      map[int]struct{}
	rGen               *rand.Rand
	nextNonRandomIndex int
	shouldSample       bool
}

func NewSampler() Sampler {
	return Sampler{
		sampleIndexes: make(map[int]struct{}),
		rGen:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

var exists = struct{}{}

func (sampler *Sampler) Reset(sliceLength int, desired int) int {

	sampler.nextNonRandomIndex = 0
	sampler.sliceLength = sliceLength
	sampler.desired = desired
	sampler.shouldSample = sliceLength > desired

	for k := range sampler.sampleIndexes {
		delete(sampler.sampleIndexes, k)
	}

	if sampler.shouldSample {
		return desired
	}
	return sliceLength
}

func (sampler *Sampler) Next() int {

	if sampler.shouldSample {

		if len(sampler.sampleIndexes) < sampler.desired {

			for {
				randomIndex := sampler.rGen.Intn(sampler.sliceLength)
				if _, ok := sampler.sampleIndexes[randomIndex]; ok {
					continue
				}
				sampler.sampleIndexes[randomIndex] = exists
				return randomIndex
			}
		}

		return -1
	}

	if sampler.nextNonRandomIndex < sampler.sliceLength {
		sampler.nextNonRandomIndex += 1
		return sampler.nextNonRandomIndex - 1
	} else {
		return -1
	}
}

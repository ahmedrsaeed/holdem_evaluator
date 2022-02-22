package slicesampler

import (
	"math/rand"
	"time"
)

type Sampler struct {
	sliceLength        int32
	desired            int
	sampleIndexes      []bool
	blank              []bool
	rGen               *rand.Rand
	nextNonRandomIndex int32
	shouldSample       bool
}

func NewSampler() Sampler {
	return Sampler{
		blank:         make([]bool, 0),
		sampleIndexes: make([]bool, 0),
		rGen:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (sampler *Sampler) Reset(sliceLength int, desired int) int {

	if sliceLength > 1<<31-1 {
		panic("can't use Rand.Int31")
	}

	sampler.nextNonRandomIndex = 0
	sampler.sliceLength = int32(sliceLength)
	sampler.shouldSample = sliceLength > desired

	if !sampler.shouldSample {
		sampler.desired = sliceLength
		return sliceLength
	}
	sampler.desired = desired

	switch {

	case len(sampler.sampleIndexes) < sliceLength:
		sampler.sampleIndexes = make([]bool, sliceLength)
	case len(sampler.blank) < len(sampler.sampleIndexes):
		sampler.blank = make([]bool, len(sampler.sampleIndexes))
		fallthrough
	default:
		copy(sampler.sampleIndexes, sampler.blank)
	}

	return desired
}

func (sampler *Sampler) Next() int32 {

	if sampler.nextNonRandomIndex < int32(sampler.desired) {
		sampler.nextNonRandomIndex += 1

		if sampler.shouldSample {
			for {
				randomIndex := sampler.rGen.Int31n(sampler.sliceLength)
				if sampler.sampleIndexes[randomIndex] {
					continue
				}
				sampler.sampleIndexes[randomIndex] = true
				return randomIndex
			}
		}
		return sampler.nextNonRandomIndex - 1
	} else {
		return -1
	}
}

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
		copy(sampler.sampleIndexes, sampler.blank[:sliceLength])
	}

	return desired
}

func (sampler *Sampler) Next() int32 {

	if sampler.nextNonRandomIndex < int32(sampler.desired) {
		sampler.nextNonRandomIndex += 1

		if sampler.shouldSample {
			for {
				randomIndex := sampler.int31n(sampler.sliceLength)
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

// int31n returns, as an int32, a non-negative pseudo-random number in the half-open interval [0,n).
// n must be > 0, but int31n does not check this; the caller must ensure it.
// int31n exists because Int31n is inefficient, but Go 1 compatibility
// requires that the stream of values produced by math/rand remain unchanged.
// int31n can thus only be used internally, by newly introduced APIs.
//
// For implementation details, see:
// https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction
// https://lemire.me/blog/2016/06/30/fast-random-shuffling
func (s *Sampler) int31n(n int32) int32 {
	v := s.rGen.Uint32()
	prod := uint64(v) * uint64(n)
	low := uint32(prod)
	if low < uint32(n) {
		thresh := uint32(-n) % uint32(n)
		for low < thresh {
			v = s.rGen.Uint32()
			prod = uint64(v) * uint64(n)
			low = uint32(prod)
		}
	}
	return int32(prod >> 32)
}

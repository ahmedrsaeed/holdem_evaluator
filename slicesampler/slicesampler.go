package slicesampler

import (
	"fmt"
	"math/bits"
	"math/rand"
	"time"
)

const maxMask uint8 = 1 << 7

type Sampler struct {
	maxSliceLength     int32
	sliceLength        int32
	sampleSize         int
	sampleIndexMask    uint8
	sampleIndexes      []uint8
	blank              []uint8
	rGen               *rand.Rand
	nextNonRandomIndex int32
	isSamplingNeeded   bool
	duplicatesFound    int
}

func NewSampler(maxSliceLength int) Sampler {
	if maxSliceLength > int(1<<31-1) {
		panic("can't use Rand.Int31")
	}
	return Sampler{
		maxSliceLength:  int32(maxSliceLength),
		sampleIndexMask: 1,
		blank:           make([]uint8, maxSliceLength),
		sampleIndexes:   make([]uint8, maxSliceLength),
		rGen:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (sampler *Sampler) Configure(sliceLength int, sampleSize int) int {
	if sliceLength > int(sampler.maxSliceLength) {
		panic(fmt.Sprintf("%d is greater than max slice length %d", sliceLength, sampler.maxSliceLength))
	}

	sampler.sliceLength = int32(sliceLength)
	sampler.isSamplingNeeded = sliceLength > sampleSize

	switch {
	case sampler.isSamplingNeeded:
		sampler.sampleSize = sampleSize
	default:
		sampler.sampleSize = sliceLength
	}

	sampler.nextNonRandomIndex = int32(sampler.sampleSize) //make it so Next will return done unless Reset is called

	return sampler.sampleSize
}

func (sampler *Sampler) Reset() {

	sampler.nextNonRandomIndex = 0

	if !sampler.isSamplingNeeded {
		return
	}

	if sampler.sampleIndexMask == maxMask {
		copy(sampler.sampleIndexes, sampler.blank)
	}
	sampler.sampleIndexMask = bits.RotateLeft8(sampler.sampleIndexMask, 1)
}

func (sampler *Sampler) Next() int32 {

	if sampler.nextNonRandomIndex < int32(sampler.sampleSize) {
		sampler.nextNonRandomIndex += 1

		if sampler.isSamplingNeeded {
			for {
				randomIndex := sampler.int31n(sampler.sliceLength)
				if sampler.sampleIndexes[randomIndex]&sampler.sampleIndexMask != 0 {
					//sampler.duplicatesFound++
					continue
				}
				sampler.sampleIndexes[randomIndex] |= sampler.sampleIndexMask
				return randomIndex
			}
		}
		return sampler.nextNonRandomIndex - 1
	} else {
		return -1
	}
}

func (sampler *Sampler) Print() {
	//fmt.Println("Duplicates found ", sampler.duplicatesFound)
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

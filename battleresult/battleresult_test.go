package battleresult

import (
	"math/rand"
	"testing"
	"time"
)

var availableglobal []int

// from battleresult.go
//go test -benchmem  -bench . -cpuprofile cpu.out
func BenchmarkAdd(b *testing.B) {
	// run the Fib function b.N times
	rand.Seed(time.Now().Unix())

	x := New()
	temp := rand.Perm(19)
	available := make([]uint8, 19)
	for i, x := range temp {
		available[i] = uint8(x)
	}

	skipIndexes := []uint8{1, 5}
	b.ResetTimer()
	for t := 0; t < b.N; t++ {
		x.Reset(17)

		for n := 0; n < 7000; n++ {
			x.Add(available, skipIndexes, n)
		}
		_, tieCount, _ := x.Next()

		availableglobal = append(availableglobal, tieCount)
	}
}

package battleresult

import (
	"bytes"
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

	br := New()
	temp := rand.Perm(19)
	available := make([]uint8, 19)
	for i, x := range temp {
		available[i] = uint8(x)
	}

	skipIndexes := []uint8{1, 5}
	b.ResetTimer()
	for t := 0; t < b.N; t++ {
		br.Configure(17)

		for n := 0; n < 7000; n++ {
			br.Add(available, skipIndexes, n)
		}
		_, tieCount, _ := br.Next()

		availableglobal = append(availableglobal, tieCount)
	}
}

func TestBattleResutl(t *testing.T) {

	br := New()
	br.Configure(4)

	br.Add([]uint8{1, 2, 3, 4, 5, 6}, []uint8{0, 1}, 2)
	br.Add([]uint8{1, 2, 3, 4, 5, 6}, []uint8{1, 2}, 3)
	br.Add([]uint8{1, 2, 3, 4, 5, 6}, []uint8{3, 4}, 4)
	br.Add([]uint8{1, 2, 3, 4, 5, 6}, []uint8{0, 5}, 5)

	testNext(t, &br, []uint8{3, 4, 5, 6}, 2, false)
	testNext(t, &br, []uint8{1, 4, 5, 6}, 3, false)
	testNext(t, &br, []uint8{1, 2, 3, 6}, 4, false)
	testNext(t, &br, []uint8{2, 3, 4, 5}, 5, false)
	testNext(t, &br, nil, 0, true)
	// expectedTC :=

}

func testNext(t *testing.T, br *BattleResult, expectedAvailable []uint8, expectedTieCount int, expectedDone bool) {

	available, tieCount, done := br.Next()

	if done != expectedDone {
		t.Errorf("expected %v got %v", expectedDone, done)
	}

	if !bytes.Equal(available, expectedAvailable) {
		t.Errorf("expected %v got %v", expectedAvailable, available)
	}

	if tieCount != expectedTieCount {
		t.Errorf("expected %v got %v", expectedTieCount, tieCount)
	}
}

package battleresult

import "fmt"

type BattleResult struct {
	backingTieCount        []int
	backingLeftOverCards   []uint8
	len                    int
	iterIndex              int
	leftOverCardsPerBattle int
}

func New() BattleResult {
	return BattleResult{
		backingLeftOverCards: make([]uint8, 0),
		backingTieCount:      make([]int, 0),
	}
}

func (br *BattleResult) Reset(leftOverCardsPerBattle int) {
	br.len = 0
	br.iterIndex = 0
	br.leftOverCardsPerBattle = leftOverCardsPerBattle
}

func (br *BattleResult) Next() ([]uint8, int, bool) {

	if br.iterIndex < br.len {
		index := br.iterIndex
		br.iterIndex++
		start := br.leftOverCardsPerBattle * index
		return br.backingLeftOverCards[start : start+br.leftOverCardsPerBattle], br.backingTieCount[index], false
	} else {
		return nil, 0, true
	}
}

func (br *BattleResult) growBackingArrays(cardsBeingAdded int) {

	if cardsBeingAdded != br.leftOverCardsPerBattle {
		panic(fmt.Sprintf("expected %d cards got %d", br.leftOverCardsPerBattle, cardsBeingAdded))
	}

	const growthFactor int = 10

	if len(br.backingTieCount) < br.len+1 {

		//fmt.Printf("Growing tie count len %d by %d for %d cards\n", len(br.backingTieCount), growthFactor, 1)
		br.backingTieCount = append(br.backingTieCount, make([]int, growthFactor)...)
	}

	if len(br.backingLeftOverCards) < (br.len+1)*cardsBeingAdded {

		//fmt.Printf("Growing left over cards results len %d by %d for %d cards\n", len(br.backingLeftOverCards), growthFactor, cardsBeingAdded)
		br.backingLeftOverCards = append(br.backingLeftOverCards, make([]uint8, growthFactor*cardsBeingAdded)...)
	}
}

func (br *BattleResult) Add(src []uint8, skipIndexes []uint8, tieCount int) {

	cardsBeingAdded := len(src) - len(skipIndexes)

	br.growBackingArrays(cardsBeingAdded)

	br.backingTieCount[br.len] = tieCount

	var dstStart = br.len * cardsBeingAdded
	var srcStart = 0
	for i := range skipIndexes {

		skipIndex := int(skipIndexes[i])
		if srcStart == skipIndex {
			srcStart++
			continue
		}

		dstEnd := dstStart + skipIndex - srcStart

		//println(dstStart, dstEnd, srcStart, skipIndex)
		copy(br.backingLeftOverCards[dstStart:dstEnd], src[srcStart:skipIndex])
		dstStart = dstEnd
		srcStart = skipIndex + 1
	}
	copy(br.backingLeftOverCards[dstStart:], src[srcStart:])
	br.len++
}

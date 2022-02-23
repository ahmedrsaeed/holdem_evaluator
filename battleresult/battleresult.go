package battleresult

type BattleResult struct {
	leftOverCards      []uint8
	leftOverCardsCount int
	tieCount           int
}

func (br *BattleResult) PairFromLeftOverCards(ia uint8, ib uint8) (uint8, uint8) {
	return br.leftOverCards[ia], br.leftOverCards[ib]
}
func (br *BattleResult) LeftOverCards() []uint8 {
	return br.leftOverCards[:br.leftOverCardsCount]
}
func (br *BattleResult) TieCount() int {
	return br.tieCount
}

type BattleResultPool struct {
	battleResultsAvailable []*BattleResult
}

func NewBattleResultPool() BattleResultPool {
	return BattleResultPool{
		battleResultsAvailable: make([]*BattleResult, 0),
	}
}

func (pool *BattleResultPool) ReturnToPool(lo *BattleResult) {
	pool.battleResultsAvailable = append(pool.battleResultsAvailable, lo)
}

func (pool *BattleResultPool) From(src []uint8, skipIndexes []uint8, tieCount int) *BattleResult {

	lastAvailableBattleIndex := len(pool.battleResultsAvailable) - 1

	var battleResult *BattleResult

	if lastAvailableBattleIndex < 0 {
		battleResult = &BattleResult{}
	} else {
		battleResult = pool.battleResultsAvailable[lastAvailableBattleIndex]
		pool.battleResultsAvailable = pool.battleResultsAvailable[:lastAvailableBattleIndex]
	}

	battleResult.leftOverCardsCount = len(src) - len(skipIndexes)
	battleResult.tieCount = tieCount

	if len(battleResult.leftOverCards) < battleResult.leftOverCardsCount {
		battleResult.leftOverCards = make([]uint8, battleResult.leftOverCardsCount)
	}

	var dstStart uint8 = 0
	var srcStart uint8 = 0
	for _, skipIndex := range skipIndexes {

		if srcStart == skipIndex {
			srcStart++
			continue
		}

		dstEnd := dstStart + skipIndex - srcStart

		//println(dstStart, dstEnd, srcStart, skipIndex)
		copy(battleResult.leftOverCards[dstStart:dstEnd], src[srcStart:skipIndex])
		dstStart = dstEnd
		srcStart = skipIndex + 1
	}
	copy(battleResult.leftOverCards[dstStart:], src[srcStart:])

	return battleResult
}

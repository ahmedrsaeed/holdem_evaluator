package battleresult

import "holdem/list"

type BattleResult struct {
	leftOverCards      []int
	leftOverCardsCount int
	tieCount           int
}

func (br *BattleResult) LeftOverCards() []int {
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

func (pool *BattleResultPool) From(src []int, indexes []int, tieCount int) *BattleResult {

	lastAvailableBattleIndex := len(pool.battleResultsAvailable) - 1

	if lastAvailableBattleIndex < 0 {

		new := BattleResult{
			leftOverCards:      list.ValuesAtIndexes(src, indexes),
			leftOverCardsCount: len(indexes),
			tieCount:           tieCount,
		}

		return &new
	}

	last := pool.battleResultsAvailable[lastAvailableBattleIndex]
	pool.battleResultsAvailable = pool.battleResultsAvailable[:lastAvailableBattleIndex]

	last.leftOverCardsCount = len(indexes)
	last.tieCount = tieCount

	if len(last.leftOverCards) < last.leftOverCardsCount {
		last.leftOverCards = list.ValuesAtIndexes(src, indexes)
	} else {
		list.CopyValuesAtIndexes(last.leftOverCards, src, indexes)
	}

	return last
}

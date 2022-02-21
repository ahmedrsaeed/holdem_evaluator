package leftovercardspool

import "holdem/list"

type LeftOverCards struct {
	store []int
	len   int
}

func (lo *LeftOverCards) Cards() []int {
	return lo.store[:lo.len]
}

type LeftOverCardsPool struct {
	leftOverCardsAvailable []*LeftOverCards
}

func NewLeftOverCardsPool() LeftOverCardsPool {
	return LeftOverCardsPool{
		leftOverCardsAvailable: make([]*LeftOverCards, 0),
	}
}

func (pool *LeftOverCardsPool) ReturnToPool(lo *LeftOverCards) {
	pool.leftOverCardsAvailable = append(pool.leftOverCardsAvailable, lo)
}

func (pool *LeftOverCardsPool) From(src []int, indexes []int) *LeftOverCards {

	lastCardIndex := len(pool.leftOverCardsAvailable) - 1

	if lastCardIndex < 0 {

		new := LeftOverCards{
			store: list.ValuesAtIndexes(src, indexes),
			len:   len(indexes),
		}

		return &new
	}

	last := pool.leftOverCardsAvailable[lastCardIndex]
	pool.leftOverCardsAvailable = pool.leftOverCardsAvailable[:lastCardIndex]

	last.len = len(indexes)

	if len(last.store) < last.len {
		last.store = list.ValuesAtIndexes(src, indexes)
	} else {
		list.CopyValuesAtIndexes(last.store, src, indexes)
	}

	return last
}

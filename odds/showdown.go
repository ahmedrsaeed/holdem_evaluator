package odds

import (
	"fmt"
	"holdem/battleresult"
	"holdem/handevaluator"
	"holdem/list"
	"holdem/slicesampler"
	"math"
)

func (calc *OddsCalculator) showDown(
	hero []uint8,
	communityKnown []uint8,
	availableToCommunity []uint8,
	villainCount int,
	desiredSamplesPerVillain int,
	communityCombinations [][]uint8,
	communityCombinationIndex <-chan int32,
	results chan<- oddsRaw) {

	//reusables need to be used immediately
	//reusableHand := make([]uint8, 2)
	reusableRemainingCommunity := make([]uint8, remainingCommunityCardsCount(communityKnown))
	battleResultPool := battleresult.NewBattleResultPool()
	comboSampler := slicesampler.NewSampler()
	lossResults := make([][]*battleresult.BattleResult, 2)
	lossResultsMaxCap := int(math.Pow(float64(desiredSamplesPerVillain), float64(villainCount)))
	lossResults[0] = make([]*battleresult.BattleResult, lossResultsMaxCap)
	lossResults[1] = make([]*battleresult.BattleResult, lossResultsMaxCap)

	rawOdds := oddsRaw{
		total:            0,
		win:              0,
		tie:              0,
		lose:             0,
		tieVillainCounts: map[int]int{},
		// villainHandsFaced:    make([]int, len(calc.allPossiblePairs)),
		// villainHandsLostTo:   make([]int, len(calc.allPossiblePairs)),
		// villainHandsTiedWith: make([]int, len(calc.allPossiblePairs)),
		hero: map[string]int{},
	}

	for communityComboIndex := range communityCombinationIndex {

		list.CopyValuesAtIndexes(reusableRemainingCommunity, availableToCommunity, communityCombinations[communityComboIndex])
		handEvaluator := calc.evaluator.CreateFrom(communityKnown, reusableRemainingCommunity)

		heroResult := handEvaluator(hero[0], hero[1])

		if heroResult.HandName == handevaluator.InvalidHand {
			panic("invalid hand for hero")
		}

		showDownsWon := 0
		showDownsTied := 0
		showDownsLost := 0
		total := 1
		previousNonLossResults := append(
			lossResults[1][:0],
			battleResultPool.From(availableToCommunity, communityCombinations[communityComboIndex], 0),
		)

		cardsAvailableToVillainCount := uint8(len(availableToCommunity) - len(communityCombinations[communityComboIndex]))
		lastVillainIndex := villainCount - 1

		for vi := 0; vi < villainCount; vi++ {

			allViCombinations, err := calc.combinations.Get(cardsAvailableToVillainCount, 2)

			if err != nil {
				panic(err.Error())
			}

			actualViSamples := comboSampler.Reset(len(allViCombinations), desiredSamplesPerVillain)
			cardsAvailableToVillainCount -= 2
			total *= actualViSamples
			showDownsLost *= actualViSamples
			currentNonLossResults := lossResults[vi%2][:0]
			for _, prev := range previousNonLossResults {

				comboSampler.Reset(len(allViCombinations), desiredSamplesPerVillain)
				for viComboIndex := comboSampler.Next(); viComboIndex > -1; viComboIndex = comboSampler.Next() {
					viCardA, viCardB := prev.PairFromLeftOverCards(
						allViCombinations[viComboIndex][0],
						allViCombinations[viComboIndex][1])
					villainResult := handEvaluator(viCardA, viCardB)

					// viKey := -1
					// if vi == 0 {
					// 	//same suit different suit bla same card
					// 	viKey = calc.allPossiblePairsIndexMap[reusableHand[0]][reusableHand[1]]
					// 	rawOdds.villainHandsFaced[viKey] += 1
					// }

					currentTieCount := prev.TieCount()
					switch {

					case villainResult.HandName == handevaluator.InvalidHand:
						panic(fmt.Sprintf("invalid hand for villain %d", vi+1))
					case villainResult.Value > heroResult.Value:
						showDownsLost++
						// if viKey > -1 {
						// 	rawOdds.villainHandsLostTo[viKey] += 1
						// }
						continue
					case villainResult.Value == heroResult.Value:
						currentTieCount++
						// if viKey > -1 {
						// 	rawOdds.villainHandsTiedWith[viKey] += 1
						// }
					default:
					}

					if vi == lastVillainIndex {

						if currentTieCount == 0 {
							showDownsWon++
						} else {
							showDownsTied++
							rawOdds.tieVillainCounts[currentTieCount] += 1
						}
						continue
					}

					//println("I should not be reached for one villain")

					currentNonLossResults = append(
						currentNonLossResults,
						battleResultPool.From(prev.LeftOverCards(), allViCombinations[viComboIndex], currentTieCount))
				}

				battleResultPool.ReturnToPool(prev)
			}
			previousNonLossResults = currentNonLossResults
		}

		rawOdds.total += total
		rawOdds.win += showDownsWon
		rawOdds.tie += showDownsTied
		rawOdds.lose += showDownsLost
		rawOdds.hero[heroResult.HandName] += total
	}

	//comboSampler.PrintDuplicateCount("go routine")
	results <- rawOdds
}

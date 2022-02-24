package odds

import (
	"fmt"
	"holdem/battleresult"
	"holdem/handevaluator"
	"holdem/list"
	"holdem/slicesampler"
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
	comboSampler := slicesampler.NewSampler()
	previousNonLossResults := battleresult.New()
	currentNonLossResults := battleresult.New()

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

		unassignedCardCount := len(availableToCommunity) - len(communityCombinations[communityComboIndex])

		previousNonLossResults.Reset(unassignedCardCount)
		previousNonLossResults.Add(availableToCommunity, communityCombinations[communityComboIndex], 0)

		lastVillainIndex := villainCount - 1

		for vi := 0; vi < villainCount; vi++ {

			allViCombinations, err := calc.combinations.Get(uint8(unassignedCardCount), 2)

			if err != nil {
				panic(err.Error())
			}

			unassignedCardCount -= 2
			currentNonLossResults.Reset(unassignedCardCount)

			actualViSamples := comboSampler.Reset(len(allViCombinations), desiredSamplesPerVillain)
			total *= actualViSamples
			showDownsLost *= actualViSamples
			for currAvailableCards, previousTieCount, done := previousNonLossResults.Next(); !done; currAvailableCards, previousTieCount, done = previousNonLossResults.Next() {

				comboSampler.Reset(len(allViCombinations), desiredSamplesPerVillain)

				for viComboIndex := comboSampler.Next(); viComboIndex > -1; viComboIndex = comboSampler.Next() {
					viCardA, viCardB := currAvailableCards[allViCombinations[viComboIndex][0]], currAvailableCards[allViCombinations[viComboIndex][1]]
					villainResult := handEvaluator(viCardA, viCardB)

					currentTieCount := previousTieCount
					switch {

					case villainResult.HandName == handevaluator.InvalidHand:
						panic(fmt.Sprintf("invalid hand for villain %d", vi+1))
					case villainResult.Value > heroResult.Value:
						showDownsLost++
						continue
					case villainResult.Value == heroResult.Value:
						currentTieCount++
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

					currentNonLossResults.Add(currAvailableCards, allViCombinations[viComboIndex], currentTieCount)
				}
			}
			previousNonLossResults, currentNonLossResults = currentNonLossResults, previousNonLossResults
		}

		rawOdds.total += total
		rawOdds.win += showDownsWon
		rawOdds.tie += showDownsTied
		rawOdds.lose += showDownsLost
		rawOdds.hero[heroResult.HandName] += total
	}

	results <- rawOdds
}

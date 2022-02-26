package odds

import (
	"fmt"
	"holdem/battleresult"
	"holdem/combinations"
	"holdem/handevaluator"
	"holdem/list"
	"holdem/slicesampler"
)

type showDownResults struct {
	total            int
	win              int
	lose             int
	tie              int
	tieVillainCounts map[int]int
	hero             []int
	// villainHandsFaced    []int
	// villainHandsLostTo   []int
	// villainHandsTiedWith []int
}

type showDown struct {
	hero                              []uint8
	communityKnown                    []uint8
	availableToCommunity              []uint8
	villainCount                      int
	desiredSamplesPerVillain          int
	communityCombinations             [][]uint8
	communityCombinationIndex         <-chan int32
	results                           chan<- showDownResults
	evaluator                         handevaluator.HandEvaluator
	combinations                      combinations.Combinations
	cardsAvailableToFirstVillainCount int
	maxVillainComboCount              int
	reusableRemainingCommunity        []uint8
	previousNonLossResults            battleresult.BattleResult
	currentNonLossResults             battleresult.BattleResult
	comboSampler                      slicesampler.Sampler
	cumulativeResults                 showDownResults
}

func (calc *OddsCalculator) showDown(
	hero []uint8,
	communityKnown []uint8,
	availableToCommunity []uint8,
	villainCount int,
	desiredSamplesPerVillain int,
	communityCombinations [][]uint8,
	communityCombinationIndex <-chan int32,
	results chan<- showDownResults) {

	showDown := showDown{
		hero:                              hero,
		communityKnown:                    communityKnown,
		availableToCommunity:              availableToCommunity,
		villainCount:                      villainCount,
		desiredSamplesPerVillain:          desiredSamplesPerVillain,
		communityCombinations:             communityCombinations,
		communityCombinationIndex:         communityCombinationIndex,
		results:                           results,
		evaluator:                         calc.evaluator,
		combinations:                      calc.combinations,
		cardsAvailableToFirstVillainCount: len(availableToCommunity) - remainingCommunityCardsCount(communityKnown),
		reusableRemainingCommunity:        make([]uint8, remainingCommunityCardsCount(communityKnown)),
		previousNonLossResults:            battleresult.New(),
		currentNonLossResults:             battleresult.New(),
		cumulativeResults: showDownResults{
			total:            0,
			win:              0,
			tie:              0,
			lose:             0,
			tieVillainCounts: map[int]int{},
			// villainHandsFaced:    make([]int, len(calc.allPossiblePairs)),
			// villainHandsLostTo:   make([]int, len(calc.allPossiblePairs)),
			// villainHandsTiedWith: make([]int, len(calc.allPossiblePairs)),
			hero: make([]int, len(handevaluator.HandTypes())),
		},
	}

	firstVillainCombinations, err := calc.combinations.Get(uint8(showDown.cardsAvailableToFirstVillainCount), 2)

	if err != nil {
		panic(err.Error())
	}

	showDown.maxVillainComboCount = len(firstVillainCombinations)
	showDown.comboSampler = slicesampler.NewSampler(showDown.maxVillainComboCount)

	for communityComboIndex := range communityCombinationIndex {
		showDown.showDownForCommunityComboIndex(communityComboIndex)
	}

	results <- showDown.cumulativeResults
}

func (sd *showDown) showDownForCommunityComboIndex(communityComboIndex int32) {

	list.CopyValuesAtIndexes(sd.reusableRemainingCommunity, sd.availableToCommunity, sd.communityCombinations[communityComboIndex])
	partialEvaluation := sd.evaluator.PartialEvaluation(sd.communityKnown, sd.reusableRemainingCommunity)

	heroValue, heroHandTypeIndex := partialEvaluation.Eval(sd.hero[0], sd.hero[1])

	if heroHandTypeIndex == handevaluator.InvalidHandIndex {
		panic("invalid hand for hero")
	}

	showDownsWon := 0
	showDownsTied := 0
	showDownsLost := 0
	total := 1

	unassignedCardCount := sd.cardsAvailableToFirstVillainCount

	sd.previousNonLossResults.Reset(unassignedCardCount)
	sd.previousNonLossResults.Add(sd.availableToCommunity, sd.communityCombinations[communityComboIndex], 0)

	lastVillainIndex := sd.villainCount - 1

	for vi := 0; vi < sd.villainCount; vi++ {

		allViCombinations, err := sd.combinations.Get(uint8(unassignedCardCount), 2)

		if err != nil {
			panic(err.Error())
		}

		unassignedCardCount -= 2
		sd.currentNonLossResults.Reset(unassignedCardCount)

		actualViSamples := sd.comboSampler.Configure(len(allViCombinations), sd.desiredSamplesPerVillain)
		total *= actualViSamples
		showDownsLost *= actualViSamples
		for currAvailableCards, previousTieCount, done := sd.previousNonLossResults.Next(); !done; currAvailableCards, previousTieCount, done = sd.previousNonLossResults.Next() {

			sd.comboSampler.Reset()

			for viComboIndex := sd.comboSampler.Next(); viComboIndex > -1; viComboIndex = sd.comboSampler.Next() {
				viCardA, viCardB := currAvailableCards[allViCombinations[viComboIndex][0]], currAvailableCards[allViCombinations[viComboIndex][1]]
				villainValue, villainHandTypeIndex := partialEvaluation.Eval(viCardA, viCardB)

				currentTieCount := previousTieCount
				switch {

				case villainHandTypeIndex == handevaluator.InvalidHandIndex:
					panic(fmt.Sprintf("invalid hand for villain %d", vi+1))
				case villainValue > heroValue:
					showDownsLost++
					continue
				case villainValue == heroValue:
					currentTieCount++
				default:
				}

				if vi == lastVillainIndex {

					if currentTieCount == 0 {
						showDownsWon++
					} else {
						showDownsTied++
						sd.cumulativeResults.tieVillainCounts[currentTieCount] += 1
					}
					continue
				}

				//println("I should not be reached for one villain")

				sd.currentNonLossResults.Add(currAvailableCards, allViCombinations[viComboIndex], currentTieCount)
			}
		}
		sd.previousNonLossResults, sd.currentNonLossResults = sd.currentNonLossResults, sd.previousNonLossResults
	}

	sd.cumulativeResults.total += total
	sd.cumulativeResults.win += showDownsWon
	sd.cumulativeResults.tie += showDownsTied
	sd.cumulativeResults.lose += showDownsLost
	sd.cumulativeResults.hero[heroHandTypeIndex] += total
}

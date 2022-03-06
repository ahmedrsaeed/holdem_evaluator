package odds

import (
	"fmt"
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

type villain struct {
	cardsAvailable []uint8
	combinations   [][]uint8
	sampler        slicesampler.Sampler
	sampleSize     int
	lossMultiplier int
	tieCount       int
}

type showDown struct {
	hero                       []uint8
	communityKnown             []uint8
	availableToCommunity       []uint8
	communityCombinations      [][]uint8
	communityCombinationIndex  <-chan int32
	results                    chan<- showDownResults
	evaluator                  handevaluator.HandEvaluator
	combinations               combinations.Combinations
	reusableRemainingCommunity []uint8
	cumulativeResults          showDownResults
	villains                   []villain
	totalPerCombo              int
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
		hero:                       hero,
		communityKnown:             communityKnown,
		availableToCommunity:       availableToCommunity,
		communityCombinations:      communityCombinations,
		communityCombinationIndex:  communityCombinationIndex,
		results:                    results,
		evaluator:                  calc.evaluator,
		combinations:               calc.combinations,
		reusableRemainingCommunity: make([]uint8, remainingCommunityCardsCount(communityKnown)),
		villains:                   make([]villain, villainCount),
		cumulativeResults: showDownResults{
			total:            0,
			win:              0,
			tie:              0,
			lose:             0,
			tieVillainCounts: map[int]int{},
			hero:             make([]int, len(handevaluator.HandTypes())),
		},
	}

	cardsAvailableToVillain := len(availableToCommunity) - remainingCommunityCardsCount(communityKnown)
	showDown.totalPerCombo = 1

	for i := range showDown.villains {
		combinations, err := calc.combinations.Get(uint8(cardsAvailableToVillain), 2)
		showDown.villains[i].cardsAvailable = make([]uint8, cardsAvailableToVillain)
		cardsAvailableToVillain -= 2

		if err != nil {
			panic(err.Error())
		}
		showDown.villains[i].combinations = combinations
		showDown.villains[i].sampler = slicesampler.NewSampler(len(combinations))
		showDown.villains[i].sampleSize = showDown.villains[i].sampler.Configure(len(combinations), desiredSamplesPerVillain)
		showDown.totalPerCombo *= showDown.villains[i].sampleSize
	}

	lossMultiplier := showDown.totalPerCombo
	for i := 0; i < villainCount; i++ {
		lossMultiplier /= showDown.villains[i].sampleSize
		showDown.villains[i].lossMultiplier = lossMultiplier
	}

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

	list.CopyValuesNotAtIndexes(sd.villains[0].cardsAvailable, sd.availableToCommunity, sd.communityCombinations[communityComboIndex])
	lastVillainIndex := len(sd.villains) - 1

	for vi := 0; vi > -1; vi-- {

		for viComboIndex := sd.villains[vi].sampler.Next(); viComboIndex > -1; viComboIndex = sd.villains[vi].sampler.Next() {
			currentViCombo := sd.villains[vi].combinations[viComboIndex]
			viCardA, viCardB := sd.villains[vi].cardsAvailable[currentViCombo[0]], sd.villains[vi].cardsAvailable[currentViCombo[1]]
			villainValue, villainHandTypeIndex := partialEvaluation.Eval(viCardA, viCardB)

			currentTieCount := sd.villains[vi].tieCount
			switch {

			case villainHandTypeIndex == handevaluator.InvalidHandIndex:
				panic(fmt.Sprintf("invalid hand for villain %d", vi+1))
			case villainValue > heroValue:
				showDownsLost += sd.villains[vi].lossMultiplier
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
			vi += 1
			list.CopyValuesNotAtIndexes(sd.villains[vi].cardsAvailable, sd.villains[vi-1].cardsAvailable, currentViCombo)
			sd.villains[vi].tieCount = currentTieCount
		}

	}

	sd.cumulativeResults.total += sd.totalPerCombo
	sd.cumulativeResults.win += showDownsWon
	sd.cumulativeResults.tie += showDownsTied
	sd.cumulativeResults.lose += showDownsLost
	sd.cumulativeResults.hero[heroHandTypeIndex] += sd.totalPerCombo
}

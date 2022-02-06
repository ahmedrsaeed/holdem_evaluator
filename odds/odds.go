package odds

import (
	"errors"
	"fmt"
	"holdem/combinations"
	"holdem/deck"
	"holdem/handevaluator"
	"holdem/list"
	"math/rand"
	"time"
)

type Outcome int64

const (
	Win Outcome = iota
	Lose
	Tie
	Invalid
)

type Odds struct {
	Total   int
	Invalid int
	Win     int
	Lose    int
	Tie     int
	Hero    map[string]int
	Villain map[string]int
}

func combinationsToCards(available []int, combinations [][]int) [][]int {
	out := make([][]int, len(combinations))

	for outIndex, combination := range combinations {

		cards := make([]int, len(combination))

		for i, n := range combination {
			cards[i] = available[n]
		}

		out[outIndex] = cards
	}

	return out
}

func sample(combinations [][]int, desired int) [][]int {

	combinationsLength := len(combinations)

	if combinationsLength <= desired {
		return combinations
	}

	sampleIndexes := make(map[int]struct{})
	exists := struct{}{}
	rGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	for len(sampleIndexes) < desired {
		sampleIndexes[rGen.Intn(combinationsLength)] = exists
	}

	sampled := make([][]int, 0, len(sampleIndexes))

	for selectedIndex := range sampleIndexes {
		sampled = append(sampled, combinations[selectedIndex])
	}

	return sampled
}

func Calculate(evaluator handevaluator.HandEvaluator, combinations combinations.Combinations,
	hero []int, community []int) (Odds, error) {

	result := Odds{
		Hero:    map[string]int{},
		Villain: map[string]int{},
	}

	communityCount := len(community)
	if len(hero) != 2 {
		return result, errors.New("please provide 2 hole cards")
	}

	if communityCount == 1 || communityCount == 2 || communityCount > 5 {
		return result, errors.New("please provide 3 or 4 or 5 community cards")
	}

	for _, handType := range handevaluator.HandTypes() {
		result.Hero[handType] = 0
		result.Villain[handType] = 0
	}

	deck := deck.AllNumberValues()
	knownToVillain := append(hero, community...)
	availableToVillain := list.Filter(deck, func(dc int) bool {
		return !list.Includes(knownToVillain, dc)
	})
	availableToVillainCount := len(availableToVillain)
	villainCombinations, err := combinations.Get(availableToVillainCount, 2)

	if err != nil {
		return result, err
	}

	villainHands := combinationsToCards(availableToVillain, villainCombinations)
	remainingCommunityCount := 5 - communityCount

	communityCombinations, err := combinations.Get(availableToVillainCount-2, remainingCommunityCount)

	if err != nil {
		return result, err
	}

	for vi, villain := range villainHands {

		availableToCommunity := list.Filter(availableToVillain, func(av int) bool {
			return !list.Includes(villain, av)
		})

		remainingCommunitiesSample := combinationsToCards(availableToCommunity, sample(communityCombinations, 5000))

		fmt.Println(vi)
		for _, remaining := range remainingCommunitiesSample {

			allCommunity := append(community, remaining...)
			heroResult := evaluator.Eval(append(allCommunity, hero...))

			villainResult := evaluator.Eval(append(allCommunity, villain...))

			switch {

			case villainResult.HandName == handevaluator.InvalidHand:
			case heroResult.HandName == handevaluator.InvalidHand:
				result.Invalid++
			case villainResult.Value < heroResult.Value:
				result.Win++
			case villainResult.Value > heroResult.Value:
				result.Lose++
			default:
				result.Tie++
			}
			result.Total++
			result.Hero[heroResult.HandName]++
			result.Villain[villainResult.HandName]++
		}

	}

	return result, nil
}

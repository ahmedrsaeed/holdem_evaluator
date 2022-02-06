package odds

import (
	"errors"
	"fmt"
	"holdem/combinations"
	"holdem/deck"
	"holdem/handevaluator"
	"holdem/list"
	"math/rand"
	"sync"
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
	WinP    float32
	LoseP   float32
	TieP    float32
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

func getSampler(combinations [][]int) func(int) [][]int {

	return func(desired int) [][]int {

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
}

func Calculate(evaluator handevaluator.HandEvaluator, combinations combinations.Combinations,
	hero []int, community []int, sampleSize int) (Odds, error) {

	handTypesMap := func() map[string]int {
		htmap := map[string]int{}

		for _, handType := range handevaluator.HandTypes() {
			htmap[handType] = 0
		}

		return htmap
	}

	resultAccumulator := Odds{
		Hero:    handTypesMap(),
		Villain: handTypesMap(),
	}

	communityCount := len(community)
	if len(hero) != 2 {
		return resultAccumulator, errors.New("please provide 2 hole cards")
	}

	if communityCount == 1 || communityCount == 2 || communityCount > 5 {
		return resultAccumulator, errors.New("please provide 3 or 4 or 5 community cards")
	}

	deck := deck.AllNumberValues()
	knownToVillain := append(hero, community...)
	availableToVillain := list.Filter(deck, func(dc int) bool {
		return !list.Includes(knownToVillain, dc)
	})
	availableToVillainCount := len(availableToVillain)
	villainCombinations, err := combinations.Get(availableToVillainCount, 2)

	if err != nil {
		return resultAccumulator, err
	}

	villainHands := combinationsToCards(availableToVillain, villainCombinations)
	remainingCommunityCount := 5 - communityCount

	communityCombinations, err := combinations.Get(availableToVillainCount-2, remainingCommunityCount)

	sampleCommunityCombinations := getSampler(communityCombinations)
	if err != nil {
		return resultAccumulator, err
	}

	results := make(chan Odds, len(villainHands))
	var wg sync.WaitGroup

	for vi, villain := range villainHands {
		wg.Add(1)
		go func(i int, govillain []int) {
			defer wg.Done()
			currentResult := Odds{
				Hero:    handTypesMap(),
				Villain: handTypesMap(),
			}
			availableToCommunity := list.Filter(availableToVillain, func(av int) bool {
				return !list.Includes(govillain, av)
			})

			remainingCommunitiesSample := combinationsToCards(availableToCommunity, sampleCommunityCombinations(sampleSize))

			for _, remaining := range remainingCommunitiesSample {

				heroResult := evaluator.Eval(community, remaining, hero)
				villainResult := evaluator.Eval(community, remaining, govillain)

				switch {

				case villainResult.HandName == handevaluator.InvalidHand:
				case heroResult.HandName == handevaluator.InvalidHand:
					currentResult.Invalid++
				case villainResult.Value < heroResult.Value:
					currentResult.Win++
				case villainResult.Value > heroResult.Value:
					currentResult.Lose++
				default:
					currentResult.Tie++
				}
				currentResult.Total++
				currentResult.Hero[heroResult.HandName]++
				currentResult.Villain[villainResult.HandName]++
			}
			results <- currentResult
		}(vi, villain)
	}

	wg.Wait()
	close(results)

	for r := range results {

		resultAccumulator.Total += r.Total
		resultAccumulator.Invalid += r.Invalid
		resultAccumulator.Win += r.Win
		resultAccumulator.Lose += r.Lose
		resultAccumulator.Tie += r.Tie

		for _, handType := range handevaluator.HandTypes() {

			resultAccumulator.Villain[handType] += r.Villain[handType]
			resultAccumulator.Hero[handType] += r.Hero[handType]
		}

	}

	resultAccumulator.WinP = float32(resultAccumulator.Win) / float32(resultAccumulator.Total)
	resultAccumulator.LoseP = float32(resultAccumulator.Lose) / float32(resultAccumulator.Total)
	resultAccumulator.TieP = float32(resultAccumulator.Tie) / float32(resultAccumulator.Total)
	fmt.Println("Odds evaluated")
	return resultAccumulator, nil
}

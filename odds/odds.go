package odds

import (
	"fmt"
	"holdem/combinations"
	"holdem/combinationssampler"
	"holdem/deck"
	"holdem/handevaluator"
	"holdem/list"
	"math"
	"sort"
	"strings"
	"sync"
)

var exists = struct{}{}

type Outcome int64

const (
	Win Outcome = iota
	Lose
	Tie
	Invalid
)

type OddsCalculator struct {
	deck         deck.Deck
	evaluator    handevaluator.HandEvaluator
	combinations combinations.Combinations
	memo         map[string]memoizedValue
	memoMutex    *sync.RWMutex
	preFlopMutex *sync.Mutex
}
type memoizedValue struct {
	result     Odds
	sampleSize int
}
type Odds struct {
	WinP             float32
	LoseP            float32
	TieP             float32
	Total            int
	Invalid          int
	Win              int
	Lose             int
	Tie              int
	TieVillainCounts map[int]int
	Hero             map[string]int
}

type battleResult struct {
	tieCount      int
	cardsLeftover []int
}

func NewCalculator(evaluator handevaluator.HandEvaluator, combinations combinations.Combinations, deck deck.Deck) OddsCalculator {
	c := OddsCalculator{
		evaluator:    evaluator,
		combinations: combinations,
		deck:         deck,
		memo:         map[string]memoizedValue{},
		memoMutex:    &sync.RWMutex{},
		preFlopMutex: &sync.Mutex{},
	}

	return c
}

func (calc *OddsCalculator) hasDuplicates(inputs ...[]int) (string, bool) {

	found := map[int]struct{}{}

	for _, cards := range inputs {
		for _, c := range cards {
			if _, ok := found[c]; ok {
				return calc.deck.NumberToString(c), true
			}
			found[c] = exists
		}
	}

	return "", false
}

func sortInts(in []int) []int {
	out := list.Clone(in)
	sort.Ints(out)
	return out
}

func (calc *OddsCalculator) getMemoKey(hero []int, community []int, villainCount int) string {

	sortedHero := sortInts(hero)
	villains := fmt.Sprintf("<-hero|%dvillains|community->", villainCount)

	if len(community) == 0 && len(hero) == 2 {

		heroValues := calc.deck.Values(sortedHero)
		if calc.deck.SameSuit(sortedHero) {
			return strings.Join(append(heroValues, "same-suit|", villains), "-")
		}
		return strings.Join(append(heroValues, "different-suit|", villains), "-")
	}

	communityStrings, err := calc.deck.NumbersToString(sortInts(community))

	if err != nil {
		return err.Error()
	}

	heroStrings, err := calc.deck.NumbersToString(sortedHero)

	if err != nil {
		return err.Error()
	}

	return strings.Join(append(
		append(heroStrings, villains),
		communityStrings...), "-")
}

func (calc *OddsCalculator) readFromMemo(key string) (memoizedValue, bool) {
	calc.memoMutex.RLock()
	defer calc.memoMutex.RUnlock()
	value, ok := calc.memo[key]
	return value, ok
}

func remainingCommunityCardsCount(communityKnown []int) int {
	return 5 - len(communityKnown)
}

func (calc *OddsCalculator) showDown(
	hero []int,
	communityKnown []int,
	availableToCommunity []int,
	villainCount int,
	desiredSamplesPerVillain int,
	communityCombination <-chan combinations.Combination,
	results chan<- Odds) {

	//reusables need to be used immediately
	reusableHand := make([]int, 2)
	reusableRemainingCommunity := make([]int, remainingCommunityCardsCount(communityKnown))

	comboSamplerCreator := combinationssampler.NewCreator()

	for communityCombo := range communityCombination {

		list.CopyValuesAtIndexes(reusableRemainingCommunity, availableToCommunity, communityCombo.Selected)
		handEvaluator := calc.evaluator.Eval(communityKnown, reusableRemainingCommunity)

		heroResult := handEvaluator(hero)

		if heroResult.HandName == handevaluator.InvalidHand {
			panic("invalid hand for hero")
		}

		wins := 0
		ties := 0
		lossCount := 0
		total := 1
		previousNonLossResults := []battleResult{{
			tieCount:      0,
			cardsLeftover: list.ValuesAtIndexes(availableToCommunity, communityCombo.Other),
		}}

		tieVillainCounts := map[int]int{}
		cardsAvailableToVillainCount := len(communityCombo.Other)
		lastVillainIndex := villainCount - 1

		for vi := 0; vi < villainCount; vi++ {

			allViCombinations, err := calc.combinations.Get(cardsAvailableToVillainCount, 2)

			if err != nil {
				panic(err.Error())
			}

			viCombinationsSampler, actualViSamples := comboSamplerCreator.Create(allViCombinations, desiredSamplesPerVillain)

			cardsAvailableToVillainCount -= 2
			total *= actualViSamples
			lossCount *= actualViSamples
			currentNonLossResults := []battleResult{}
			tieVillainCounts[vi+1] = 0
			for _, prev := range previousNonLossResults {

				viCombinationsSampler(func(viCombo combinations.Combination) {

					list.CopyValuesAtIndexes(reusableHand, prev.cardsLeftover, viCombo.Selected)
					villainResult := handEvaluator(reusableHand)

					tieCount := prev.tieCount
					switch {

					case villainResult.HandName == handevaluator.InvalidHand:
						panic(fmt.Sprintf("invalid hand for villain %d", vi+1))
					case villainResult.Value > heroResult.Value:
						lossCount++
						return
					case villainResult.Value == heroResult.Value:
						tieCount++
					default:
					}

					if vi == lastVillainIndex {

						if tieCount == 0 {
							wins++
						} else {
							ties++
							tieVillainCounts[tieCount] += 1
						}
						return
					}

					//println("I should not be reached for one villain")

					currentNonLossResults = append(currentNonLossResults, battleResult{
						tieCount:      tieCount,
						cardsLeftover: list.ValuesAtIndexes(prev.cardsLeftover, viCombo.Other),
					})
				})
			}
			previousNonLossResults = currentNonLossResults
		}

		results <- Odds{
			Total:            total,
			Win:              wins,
			Tie:              ties,
			Lose:             lossCount,
			TieVillainCounts: tieVillainCounts,
			Hero:             map[string]int{heroResult.HandName: total},
		}
	}
}

func handTypesMap() map[string]int {
	htmap := map[string]int{}

	for _, handType := range handevaluator.HandTypes() {
		htmap[handType] = 0
	}

	return htmap
}

func (calc *OddsCalculator) Calculate(heroStrings []string, communityStrings []string, villainCount int) (Odds, error) {

	resultAccumulator := Odds{
		Hero: handTypesMap(),
	}

	if villainCount < 1 || villainCount > 9 {
		return resultAccumulator, fmt.Errorf("between 1 and 9 villains supported")
	}
	// const maxSampleSize = 100000
	// const minSampleSize = 1000
	// if sampleSize < minSampleSize || sampleSize > maxSampleSize {
	// 	return resultAccumulator, fmt.Errorf("sample size between %d and %d is allowed", minSampleSize, maxSampleSize)
	// }

	hero, err := calc.deck.CardStringsToNumbers(heroStrings)

	if err != nil {
		return resultAccumulator, err
	}

	community, err := calc.deck.CardStringsToNumbers(communityStrings)

	if err != nil {
		return resultAccumulator, err
	}

	if len(hero) != 2 {
		return resultAccumulator, fmt.Errorf("please provide 2 hole cards")
	}

	acceptedCommunityCount := map[int]struct{}{
		0: exists,
		3: exists,
		4: exists,
		5: exists,
	}
	if _, ok := acceptedCommunityCount[len(community)]; !ok {
		return resultAccumulator, fmt.Errorf("please provide 0 or 3 or 4 or 5 community cards")
	}

	if duplicate, found := calc.hasDuplicates(hero, community); found {
		return resultAccumulator, fmt.Errorf("found more than one " + duplicate)
	}

	// memoKey := calc.getMemoKey(hero, community, villainCount)
	// fmt.Println("Memo Key: " + memoKey)

	// if cached, ok := calc.readFromMemo(memoKey); ok && cached.sampleSize >= sampleSize {
	// 	fmt.Println("Serving cached")
	// 	return cached.result, nil
	// }

	// if communityCount == 0 && sampleSize > 5000 {
	// 	fmt.Println("Waiting to compute expensive preflop " + memoKey)
	// 	calc.preFlopMutex.Lock()
	// 	defer calc.preFlopMutex.Unlock()

	// 	if cached, ok := calc.readFromMemo(memoKey); ok && cached.sampleSize >= sampleSize {
	// 		fmt.Println("expensive preflop now in cache " + memoKey)
	// 		return cached.result, nil
	// 	}
	// }

	deck := calc.deck.AllNumberValues()
	knownToCommunity := append(hero, community...)
	availableToCommunity := list.Filter(deck, func(dc int) bool {
		return !list.Includes(knownToCommunity, dc)
	})
	availableToCommunityCount := len(availableToCommunity)
	remainingCommunityCount := remainingCommunityCardsCount(community)
	//

	combinationsSamplerCreator := combinationssampler.NewCreator()

	allRemainingCommunityCombinations, err := calc.combinations.Get(availableToCommunityCount, remainingCommunityCount)

	if err != nil {
		return resultAccumulator, err
	}

	communityCombosSamplesDesiredCount := 1000 * 1000
	totalTestsDesired := float64(communityCombosSamplesDesiredCount) * 1000.0
	_, actualCommunityCombosSampleCount := combinationsSamplerCreator.Create(allRemainingCommunityCombinations, communityCombosSamplesDesiredCount)

	desiredSamplesPerVillain := int(math.Pow(totalTestsDesired/float64(actualCommunityCombosSampleCount), 1.0/float64(villainCount)))
	fmt.Printf("Desired Samples Per Villain %d\n", desiredSamplesPerVillain)

	desiredCommunityCombinationsCount := totalTestsDesired / math.Pow(float64(desiredSamplesPerVillain), float64(villainCount))

	communityCombinationsSampler, actualCommunityCombosSampleCount := combinationsSamplerCreator.Create(allRemainingCommunityCombinations, int(desiredCommunityCombinationsCount))
	fmt.Printf("Community combinations count %d\n", actualCommunityCombosSampleCount)

	if err != nil {
		return resultAccumulator, err
	}

	remainingCommuntiyCombinationsChannel := make(chan combinations.Combination, actualCommunityCombosSampleCount)
	results := make(chan Odds, actualCommunityCombosSampleCount)

	for w := 0; w < 100; w++ {
		go calc.showDown(hero, community, availableToCommunity, villainCount, desiredSamplesPerVillain, remainingCommuntiyCombinationsChannel, results)
	}

	communityCombinationsSampler(func(remainingCombo combinations.Combination) {
		remainingCommuntiyCombinationsChannel <- remainingCombo
	})

	close(remainingCommuntiyCombinationsChannel)

	fmt.Println("closed communtiyCombinationsChannel")

	resultAccumulator.TieVillainCounts = map[int]int{}
	for i := 0; i < actualCommunityCombosSampleCount; i++ {

		r := <-results
		resultAccumulator.Total += r.Total
		resultAccumulator.Invalid += r.Invalid
		resultAccumulator.Win += r.Win
		resultAccumulator.Lose += r.Lose
		resultAccumulator.Tie += r.Tie

		// if resultAccumulator.Total%100000 == 0 {
		// 	fmt.Println(resultAccumulator.Total)
		// }
		for _, handType := range handevaluator.HandTypes() {

			resultAccumulator.Hero[handType] += r.Hero[handType]
		}

		for k, tieCounts := range r.TieVillainCounts {

			resultAccumulator.TieVillainCounts[k] += tieCounts
		}

	}
	resultAccumulator.WinP = 100 * float32(resultAccumulator.Win) / float32(resultAccumulator.Total)
	resultAccumulator.LoseP = 100 * float32(resultAccumulator.Lose) / float32(resultAccumulator.Total)
	resultAccumulator.TieP = 100 * float32(resultAccumulator.Tie) / float32(resultAccumulator.Total)
	fmt.Println("Odds evaluated")

	// calc.memoMutex.Lock()
	// defer calc.memoMutex.Unlock()
	// if cached, ok := calc.memo[memoKey]; ok && cached.result.Total > resultAccumulator.Total {
	// 	fmt.Println("result discard " + memoKey)
	// 	return cached.result, nil
	// }

	// calc.memo[memoKey] = memoizedValue{
	// 	result:     resultAccumulator,
	// 	sampleSize: sampleSize,
	// }

	return resultAccumulator, nil
}

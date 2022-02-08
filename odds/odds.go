package odds

import (
	"errors"
	"fmt"
	"holdem/combinations"
	"holdem/deck"
	"holdem/handevaluator"
	"holdem/list"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
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

func (calc *OddsCalculator) getCombinationsSampler(n int, r int) (func(int) [][]int, error) {

	combinations, err := calc.combinations.Get(n, r)

	if err != nil {
		return nil, err
	}

	return func(desired int) [][]int {

		combinationsLength := len(combinations)

		if combinationsLength <= desired {
			return combinations
		}

		sampleIndexes := make(map[int]struct{})
		rGen := rand.New(rand.NewSource(time.Now().UnixNano()))

		for len(sampleIndexes) < desired {
			sampleIndexes[rGen.Intn(combinationsLength)] = exists
		}

		sampled := make([][]int, 0, len(sampleIndexes))

		for selectedIndex := range sampleIndexes {
			sampled = append(sampled, combinations[selectedIndex])
		}

		return sampled
	}, nil
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

func clone(in []int) []int {
	out := make([]int, len(in))
	copy(out, in)
	return out
}

func sortInts(in []int) []int {
	out := clone(in)
	sort.Ints(out)
	return out
}

func (calc *OddsCalculator) getMemoKey(hero []int, community []int) string {

	sortedHero := sortInts(hero)

	if len(community) == 0 && len(hero) == 2 {

		heroValues := calc.deck.Values(sortedHero)
		if calc.deck.SameSuit(sortedHero) {
			return strings.Join(append(heroValues, "same-suit"), "-")
		}
		return strings.Join(append(heroValues, "different-suit"), "-")
	}

	communityStrings, err := calc.deck.NumbersToString(sortInts(community))

	if err != nil {
		return err.Error()
	}

	heroStrings, err := calc.deck.NumbersToString(sortedHero)

	if err != nil {
		return err.Error()
	}

	return strings.Join(append(append(heroStrings, "<-hero|community->"), communityStrings...), "-")
}

func (calc *OddsCalculator) readFromMemo(key string) (memoizedValue, bool) {
	calc.memoMutex.RLock()
	defer calc.memoMutex.RUnlock()
	value, ok := calc.memo[key]
	return value, ok
}

func (calc *OddsCalculator) Calculate(heroStrings []string, communityStrings []string, sampleSize int) (Odds, error) {

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

	const maxSampleSize = 100000
	const minSampleSize = 1000
	if sampleSize < minSampleSize || sampleSize > maxSampleSize {
		return resultAccumulator, fmt.Errorf("sample size between %d and %d is allowed", minSampleSize, maxSampleSize)
	}

	hero, err := calc.deck.CardStringsToNumbers(heroStrings)

	if err != nil {
		return resultAccumulator, err
	}

	community, err := calc.deck.CardStringsToNumbers(communityStrings)

	if err != nil {
		return resultAccumulator, err
	}

	communityCount := len(community)
	if len(hero) != 2 {
		return resultAccumulator, errors.New("please provide 2 hole cards")
	}

	acceptedCommunityCount := map[int]struct{}{
		0: exists,
		3: exists,
		4: exists,
		5: exists,
	}
	if _, ok := acceptedCommunityCount[communityCount]; !ok {
		return resultAccumulator, errors.New("please provide 0 or 3 or 4 or 5 community cards")
	}

	if duplicate, found := calc.hasDuplicates(hero, community); found {
		return resultAccumulator, errors.New("found more than one " + duplicate)
	}

	memoKey := calc.getMemoKey(hero, community)
	fmt.Println("Memo Key: " + memoKey)

	if cached, ok := calc.readFromMemo(memoKey); ok && cached.sampleSize >= sampleSize {
		fmt.Println("Serving cached")
		return cached.result, nil
	}

	if communityCount == 0 && sampleSize > 5000 {
		fmt.Println("Waiting to compute expensive preflop " + memoKey)
		calc.preFlopMutex.Lock()
		defer calc.preFlopMutex.Unlock()

		if cached, ok := calc.readFromMemo(memoKey); ok && cached.sampleSize >= sampleSize {
			fmt.Println("expensive preflop now in cache " + memoKey)
			return cached.result, nil
		}
	}

	deck := calc.deck.AllNumberValues()
	knownToVillain := append(hero, community...)
	availableToVillain := list.Filter(deck, func(dc int) bool {
		return !list.Includes(knownToVillain, dc)
	})
	availableToVillainCount := len(availableToVillain)
	villainCombinations, err := calc.combinations.Get(availableToVillainCount, 2)

	if err != nil {
		return resultAccumulator, err
	}

	villainHands := combinationsToCards(availableToVillain, villainCombinations)
	remainingCommunityCount := 5 - communityCount

	sampleCommunityCombinations, err := calc.getCombinationsSampler(availableToVillainCount-2, remainingCommunityCount)

	if err != nil {
		return resultAccumulator, err
	}

	results := make(chan Odds, len(villainHands))
	var wg sync.WaitGroup

	heroEvaluator := calc.evaluator.Eval(hero, community)

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
			villainEvaluator := calc.evaluator.Eval(govillain, community)

			remainingCommunitiesSample := combinationsToCards(availableToCommunity, sampleCommunityCombinations(sampleSize))

			for _, remaining := range remainingCommunitiesSample {

				heroResult := heroEvaluator(remaining)
				villainResult := villainEvaluator(remaining)

				switch {

				case villainResult.HandName == handevaluator.InvalidHand:
					fallthrough
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

	calc.memoMutex.Lock()
	defer calc.memoMutex.Unlock()
	if cached, ok := calc.memo[memoKey]; ok && cached.result.Total > resultAccumulator.Total {
		fmt.Println("result discard " + memoKey)
		return cached.result, nil
	}

	calc.memo[memoKey] = memoizedValue{
		result:     resultAccumulator,
		sampleSize: sampleSize,
	}

	return resultAccumulator, nil
}

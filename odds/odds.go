package odds

import (
	"fmt"
	"holdem/battleresult"
	"holdem/combinations"
	"holdem/combinationssampler"
	"holdem/deck"
	"holdem/handevaluator"
	"holdem/list"
	"math"
	"runtime"
	"sort"
	"strings"
	"sync"
)

var exists = struct{}{}

type OddsCalculator struct {
	deck                     deck.Deck
	evaluator                handevaluator.HandEvaluator
	combinations             combinations.Combinations
	memo                     map[string]memoizedValue
	memoMutex                *sync.RWMutex
	preFlopMutex             *sync.Mutex
	allPossiblePairs         [][]string
	allPossiblePairsIndexMap map[int]map[int]int
}
type memoizedValue struct {
	result     Odds
	sampleSize int
}

type HandComparision struct {
	Hand          []string
	BeatHeroP     float32
	TiedWithHeroP float32
	Total         int
}

type Probabilities struct {
	Win  float32
	Lose float32
	Tie  float32
}
type Totals struct {
	Total int
	Win   int
	Lose  int
	Tie   int
}
type Odds struct {
	Probabilities    Probabilities
	Totals           Totals
	TieVillainCounts map[int]int
	Hero             map[string]int
	// HandComparisions []HandComparision
}

type oddsRaw struct {
	total            int
	win              int
	lose             int
	tie              int
	tieVillainCounts map[int]int
	hero             map[string]int
	// villainHandsFaced    []int
	// villainHandsLostTo   []int
	// villainHandsTiedWith []int
}

func NewCalculator(evaluator handevaluator.HandEvaluator, combinations combinations.Combinations, deck deck.Deck) OddsCalculator {

	allPossibleNumberPairs, allPossiblePairsIndexMap, err := combinations.GetAllPossiblePairs(deck.AllNumberValues())

	if err != nil {
		panic(err)
	}

	allPossibleStringPairs := make([][]string, len(allPossibleNumberPairs))

	for i, nPair := range allPossibleNumberPairs {

		res, err := deck.CardNumbersToStrings(nPair)
		if err != nil {
			panic(err)
		}
		allPossibleStringPairs[i] = res
	}

	// for i, x := range allPossibleNumberPairs {

	// 	fmt.Printf("%d %v %d %v\n", i, x, allPossiblePairsIndexMap[x[0]][x[1]], allPossibleStringPairs[i])
	// }

	c := OddsCalculator{
		evaluator:                evaluator,
		combinations:             combinations,
		deck:                     deck,
		memo:                     map[string]memoizedValue{},
		memoMutex:                &sync.RWMutex{},
		preFlopMutex:             &sync.Mutex{},
		allPossiblePairs:         allPossibleStringPairs,
		allPossiblePairsIndexMap: allPossiblePairsIndexMap,
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

		heroValues, err := calc.deck.CardNumbersToStrings(sortedHero)
		if err != nil {
			return err.Error()
		}

		if calc.deck.SameSuit(sortedHero) {
			return strings.Join(append(heroValues, "same-suit|", villains), "-")
		}
		return strings.Join(append(heroValues, "different-suit|", villains), "-")
	}

	communityStrings, err := calc.deck.CardNumbersToStrings(sortInts(community))

	if err != nil {
		return err.Error()
	}

	heroStrings, err := calc.deck.CardNumbersToStrings(sortedHero)

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
	communityCombinations []combinations.Combination,
	communityCombinationIndex <-chan int,
	results chan<- oddsRaw) {

	//reusables need to be used immediately
	reusableHand := make([]int, 2)
	reusableRemainingCommunity := make([]int, remainingCommunityCardsCount(communityKnown))
	battleResultPool := battleresult.NewBattleResultPool()
	comboSampler := combinationssampler.NewSampler()
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

		list.CopyValuesAtIndexes(reusableRemainingCommunity, availableToCommunity, communityCombinations[communityComboIndex].Selected)
		handEvaluator := calc.evaluator.Eval(communityKnown, reusableRemainingCommunity)

		heroResult := handEvaluator(hero)

		if heroResult.HandName == handevaluator.InvalidHand {
			panic("invalid hand for hero")
		}

		showDownsWon := 0
		showDownsTied := 0
		showDownsLost := 0
		total := 1
		previousNonLossResults := append(
			lossResults[1][:0],
			battleResultPool.From(availableToCommunity, communityCombinations[communityComboIndex].Other, 0),
		)

		cardsAvailableToVillainCount := len(communityCombinations[communityComboIndex].Other)
		lastVillainIndex := villainCount - 1

		for vi := 0; vi < villainCount; vi++ {

			allViCombinations, err := calc.combinations.Get(cardsAvailableToVillainCount, 2)

			if err != nil {
				panic(err.Error())
			}

			actualViSamples := comboSampler.Setup(len(allViCombinations), desiredSamplesPerVillain)
			cardsAvailableToVillainCount -= 2
			total *= actualViSamples
			showDownsLost *= actualViSamples
			currentNonLossResults := lossResults[vi%2][:0]
			for _, prev := range previousNonLossResults {

				comboSampler.Setup(len(allViCombinations), desiredSamplesPerVillain)
				for viComboIndex := comboSampler.Next(); viComboIndex > -1; viComboIndex = comboSampler.Next() {
					list.CopyValuesAtIndexes(reusableHand, prev.LeftOverCards(), allViCombinations[viComboIndex].Selected)
					villainResult := handEvaluator(reusableHand)

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
						battleResultPool.From(prev.LeftOverCards(), allViCombinations[viComboIndex].Other, currentTieCount))
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

	results <- rawOdds
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

	combinationsSampler := combinationssampler.NewSampler()

	allRemainingCommunityCombinations, err := calc.combinations.Get(availableToCommunityCount, remainingCommunityCount)

	if err != nil {
		return resultAccumulator, err
	}

	communityCombosSamplesTargetCount := 300 * 1000
	totalTestsDesired := float64(communityCombosSamplesTargetCount) * 3000.0
	actualCommunityCombosSampleCount := combinationsSampler.Setup(len(allRemainingCommunityCombinations), communityCombosSamplesTargetCount)

	desiredSamplesPerVillain := int(math.Pow(totalTestsDesired/float64(actualCommunityCombosSampleCount), 1.0/float64(villainCount)))
	communityCombinationsReadjustedTargetCount := totalTestsDesired / math.Pow(float64(desiredSamplesPerVillain), float64(villainCount))
	fmt.Printf("%d villains\n", villainCount)
	fmt.Printf("Desired Samples Per Villain %d\n", desiredSamplesPerVillain)

	actualCommunityCombosSampleReadjustedCount := combinationsSampler.Setup(len(allRemainingCommunityCombinations), int(communityCombinationsReadjustedTargetCount))
	fmt.Printf("Community combinations count %d\n", actualCommunityCombosSampleReadjustedCount)

	if err != nil {
		return resultAccumulator, err
	}

	remainingCommuntiyCombinationsIndexChannel := make(chan int, actualCommunityCombosSampleReadjustedCount)
	workerCount := runtime.NumCPU()
	results := make(chan oddsRaw, workerCount)

	fmt.Printf("Worker count: %d\n", workerCount)

	for w := 0; w < workerCount; w++ {
		go calc.showDown(hero, community, availableToCommunity, villainCount, desiredSamplesPerVillain,
			allRemainingCommunityCombinations, remainingCommuntiyCombinationsIndexChannel, results)
	}

	for index := combinationsSampler.Next(); index > -1; index = combinationsSampler.Next() {
		remainingCommuntiyCombinationsIndexChannel <- index
	}

	close(remainingCommuntiyCombinationsIndexChannel)

	fmt.Println("closed communtiyCombinationsChannel")

	resultAccumulator.TieVillainCounts = map[int]int{}

	// villainHandsFaced := make([]int, len(calc.allPossiblePairs))
	// villainHandsLostTo := make([]int, len(calc.allPossiblePairs))
	// villainHandsTiedWith := make([]int, len(calc.allPossiblePairs))

	for i := 0; i < workerCount; i++ {

		r := <-results
		resultAccumulator.Totals.Total += r.total
		resultAccumulator.Totals.Win += r.win
		resultAccumulator.Totals.Lose += r.lose
		resultAccumulator.Totals.Tie += r.tie

		// if resultAccumulator.Total%100000 == 0 {
		// 	fmt.Println(resultAccumulator.Total)
		// }
		for _, handType := range handevaluator.HandTypes() {

			resultAccumulator.Hero[handType] += r.hero[handType]
		}

		for k, count := range r.tieVillainCounts {

			resultAccumulator.TieVillainCounts[k] += count
		}

		// for k, count := range r.villainHandsFaced {

		// 	villainHandsFaced[k] += count
		// }
		// for k, count := range r.villainHandsLostTo {

		// 	villainHandsLostTo[k] += count
		// }

		// for k, count := range r.villainHandsTiedWith {

		// 	villainHandsTiedWith[k] += count
		// }

	}
	resultAccumulator.Probabilities.Win = 100 * float32(resultAccumulator.Totals.Win) / float32(resultAccumulator.Totals.Total)
	resultAccumulator.Probabilities.Lose = 100 * float32(resultAccumulator.Totals.Lose) / float32(resultAccumulator.Totals.Total)
	resultAccumulator.Probabilities.Tie = 100 * float32(resultAccumulator.Totals.Tie) / float32(resultAccumulator.Totals.Total)

	// resultAccumulator.HandComparisions = make([]HandComparision, 0)
	// for k, handsFaced := range villainHandsFaced {
	// 	if handsFaced == 0 {
	// 		continue
	// 	}

	// 	resultAccumulator.HandComparisions = append(resultAccumulator.HandComparisions, HandComparision{
	// 		Hand:          calc.allPossiblePairs[k],
	// 		Total:         handsFaced,
	// 		BeatHeroP:     100 * float32(villainHandsLostTo[k]) / float32(handsFaced),
	// 		TiedWithHeroP: 100 * float32(villainHandsTiedWith[k]) / float32(handsFaced),
	// 	})
	// 	sort.Slice(resultAccumulator.HandComparisions, func(i, j int) bool {
	// 		return resultAccumulator.HandComparisions[i].BeatHeroP > resultAccumulator.HandComparisions[j].BeatHeroP
	// 	})
	// }
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

package main

import (
	"encoding/json"
	"fmt"
	"holdem/combinations"
	"holdem/deck"
	"holdem/handevaluator"
	"holdem/odds"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
)

type patternHandler struct {
	pattern string
	handler func(w http.ResponseWriter, r *http.Request)
}

func badRequest(w http.ResponseWriter, m string) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"message": m})
}

func caselessMatcher(handlers []patternHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")
		path := strings.ToLower(r.URL.Path)

		for _, h := range handlers {
			if strings.ToLower(h.pattern) == path {
				h.handler(w, r)
				return
			}
		}

		http.NotFound(w, r)
	}
}

func handleRequests() {
	evaluator, err := handevaluator.New()

	if err != nil {
		fmt.Println(err)
		return
	}

	deck := deck.New()
	oddsCalculator := odds.NewCalculator(evaluator, combinations.New(), deck)

	http.HandleFunc("/", caselessMatcher([]patternHandler{
		{pattern: "/evaluatehand", handler: getHandEvaluator(evaluator, deck)},
		{pattern: "/evaluateodds", handler: getOddsEvaluator(oddsCalculator)},
		//{pattern: "/generatecombinations", handler: getCombinationsGenerator()},
		//{pattern: "/generatepairs", handler: getPairsGenerator()},
	}))

	port := ":8081"
	fmt.Println("Preparing to listen on port " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// func getPairsGenerator() func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, req *http.Request) {

// 		n, err := iQueryParam(req, "n", 2)

// 		if err != nil {
// 			badRequest(w, err.Error())
// 			return
// 		}

// 		if n < 2 || n%2 != 0 {
// 			badRequest(w, "Please select an even n greater than 0")
// 			return
// 		}

// 		pairs := combinations.GeneratePairs(n)

// 		//fmt.Printf("%d %d %v", n, r, combinations)

// 		json.NewEncoder(w).Encode(pairs)
// 	}
// }
// func getCombinationsGenerator() func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, req *http.Request) {

// 		n, err := iQueryParam(req, "n", 0)

// 		if err != nil {
// 			badRequest(w, err.Error())
// 			return
// 		}

// 		r, err := iQueryParam(req, "r", 0)

// 		if err != nil {
// 			badRequest(w, err.Error())
// 			return
// 		}

// 		_, combinations := combinations.Generate(n, r)

// 		//fmt.Printf("%d %d %v", n, r, combinations)

// 		json.NewEncoder(w).Encode(combinations)
// 	}
// }
func getHandEvaluator(evaluator handevaluator.HandEvaluator, deck deck.Deck) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Endpoint Hit: evaluate hand")

		cards := r.URL.Query()["c"]

		if len(cards) != 7 {
			badRequest(w, "Please provide 7 cards")
			return
		}

		hand, err := deck.CardStringsToNumbers(cards)

		if err != nil {
			badRequest(w, err.Error())
			return
		}

		partialEvaluation := evaluator.PartialEvaluation(hand[:5])
		value, handTypeIndex := partialEvaluation.Eval(hand[5], hand[6])

		json.NewEncoder(w).Encode(handevaluator.EvaluatedHand{
			Value:    value,
			HandName: handevaluator.HandTypes()[handTypeIndex],
		})
	}
}

func getOddsEvaluator(oddsCalculator odds.OddsCalculator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Endpoint Hit: evaluate odds")

		community := r.URL.Query()["community"]
		hero := r.URL.Query()["hero"]

		// sampleSize, err := iQueryParam(r, "size", 100000)

		// if err != nil {
		// 	badRequest(w, err.Error())
		// 	return
		// }

		villainCount, err := iQueryParam(r, "villaincount", 1)

		if err != nil {
			badRequest(w, err.Error())
			return
		}

		result, err := oddsCalculator.Calculate(hero, community, villainCount)

		if err != nil {
			badRequest(w, err.Error())
			return
		}

		json.NewEncoder(w).Encode(result)
	}
}

func iQueryParam(r *http.Request, key string, defaultValue int) (int, error) {

	values := r.URL.Query()[key]
	if len(values) == 0 {
		return defaultValue, nil
	}
	if len(values) > 1 {
		return 0, fmt.Errorf("send only one " + key + " per call")
	}

	return strconv.Atoi(values[0])
}

func main() {
	handleRequests()
}

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
	// hand := []int{5, 10, 42, 44, 52, 2, 3}
	// hand1 := []int{26, 28, 35, 47, 2, 3, 29}

	combinations := combinations.New()
	evaluator, err := handevaluator.New()

	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/", caselessMatcher([]patternHandler{
		{pattern: "/evaluatehand", handler: getHandEvaluator(evaluator)},
		{pattern: "/evaluateodds", handler: getOddsEvaluator(evaluator, combinations)},
	}))

	port := ":8081"
	fmt.Println("Preparing to listen on port " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func getHandEvaluator(evaluator handevaluator.HandEvaluator) func(w http.ResponseWriter, r *http.Request) {
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

		json.NewEncoder(w).Encode(evaluator.Eval(hand))
	}
}

func getOddsEvaluator(evaluator handevaluator.HandEvaluator, combinations combinations.Combinations) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Endpoint Hit: evaluate odds")

		community := r.URL.Query()["community"]
		hero := r.URL.Query()["hero"]
		size := r.URL.Query()["size"]

		heroN, err := deck.CardStringsToNumbers(hero)

		if err != nil {
			badRequest(w, err.Error())
			return
		}

		communityN, err := deck.CardStringsToNumbers(community)

		if err != nil {
			badRequest(w, err.Error())
			return
		}

		sizeString := "100000"
		if len(size) > 1 {
			badRequest(w, "send only one size per call")
			return
		} else if len(size) == 1 {
			sizeString = size[0]
		}

		sampleSize, err := strconv.Atoi(sizeString)

		if err != nil {
			badRequest(w, err.Error())
			return
		}

		result, err := odds.Calculate(evaluator, combinations, heroN, communityN, sampleSize)

		if err != nil {
			badRequest(w, err.Error())
			return
		}

		json.NewEncoder(w).Encode(result)
	}
}

func main() {
	handleRequests()
}

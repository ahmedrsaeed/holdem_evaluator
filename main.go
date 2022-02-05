package main

import (
	"encoding/json"
	"fmt"
	"holdem/combinations"
	"holdem/deck"
	"holdem/handevaluator"
	"log"
	"net/http"
	"strings"
)

type urlHandler struct {
	url     string
	handler func(w http.ResponseWriter, r *http.Request)
}

func badRequest(w http.ResponseWriter, m string) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"message": m})
}

func caselessMatcher(handlers []urlHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		url := strings.ToLower(r.URL.Path)

		for _, h := range handlers {
			if h.url == url {
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

	_ = combinations.New()
	evaluator, err := handevaluator.New()

	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/", caselessMatcher([]urlHandler{
		{url: "/evaluatehand", handler: getHandEvaluator(evaluator)},
		{url: "/evaluateodds", handler: getOddsEvaluator(evaluator)},
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

func getOddsEvaluator(evaluator handevaluator.HandEvaluator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Endpoint Hit: evaluate odds")

		community := r.URL.Query()["community"]
		hero := r.URL.Query()["hero"]

		communityCount := len(community)
		if len(hero) != 2 {
			badRequest(w, "Please provide 2 hole cards.")
			return
		}

		if communityCount == 1 || communityCount == 2 || communityCount > 5 {
			badRequest(w, "Please provide 3, 4 or 5 community cards.")
			return
		}

		_, err := deck.CardStringsToNumbers(hero)

		if err != nil {
			badRequest(w, err.Error())
			return
		}

		_, err = deck.CardStringsToNumbers(community)

		if err != nil {
			badRequest(w, err.Error())
			return
		}

		json.NewEncoder(w).Encode("not implemented")
	}
}

func main() {
	handleRequests()
}

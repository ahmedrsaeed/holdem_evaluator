package main

import (
	"encoding/json"
	"fmt"
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
	fmt.Fprint(w, m)
}

func getHandEvaluator(evaluator handevaluator.HandEvaluator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Endpoint Hit: evaluate hand")

		q := r.URL.Query()["c"]

		if len(q) != 7 {
			badRequest(w, "Please provide 7 cards")
			return
		}

		hand, err := deck.CardStringsToNumbers(q)

		if err != nil {
			badRequest(w, err.Error())
			return
		}

		json.NewEncoder(w).Encode(evaluator.Eval(hand))
	}
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

	http.HandleFunc("/", caselessMatcher([]urlHandler{
		{url: "/evaluatehand", handler: getHandEvaluator(handevaluator.New())},
	}))

	log.Fatal(http.ListenAndServe(":8081", nil))

}

func main() {
	handleRequests()
}

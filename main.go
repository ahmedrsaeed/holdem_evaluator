package main

import (
	"encoding/json"
	"fmt"
	"holdem/handevaluator"
	"log"
	"net/http"
	"strings"
)

type Article struct {
	Title   string `json:"Title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

type urlHandler struct {
	url     string
	handler func(w http.ResponseWriter, r *http.Request)
}

func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllArticles")
	Articles := []Article{
		{Title: "Hello", Desc: "Article Description", Content: "Article Content"},
		{Title: "Hello 2", Desc: "Article Description", Content: "Article Content"},
	}
	json.NewEncoder(w).Encode(Articles)
}

func getHandEvaluator(evaluator handevaluator.HandEvaluator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Endpoint Hit: evaluate hand")

		q := r.URL.Query()["c"]

		if len(q) != 7 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Please provide 7 cards")
			return
		}

		json.NewEncoder(w).Encode(evaluator.Eval([]int{5, 10, 42, 44, 52, 2, 3}))
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
		{url: "/articles", handler: returnAllArticles},
	}))

	log.Fatal(http.ListenAndServe(":8081", nil))

}

func main() {
	handleRequests()
}

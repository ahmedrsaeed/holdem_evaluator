package deck

import (
	"errors"
	"fmt"
	"strings"
)

type stringNumberPair struct {
	str    string
	number uint8
}

type Deck struct {
	stringNumberPair []stringNumberPair
	stringToNumber   map[string]uint8
	numberToString   map[uint8]string
}

func New() Deck {
	deck := Deck{
		stringToNumber: map[string]uint8{},
		numberToString: map[uint8]string{},
		stringNumberPair: []stringNumberPair{
			{"2c", 1},
			{"2d", 2},
			{"2h", 3},
			{"2s", 4},
			{"3c", 5},
			{"3d", 6},
			{"3h", 7},
			{"3s", 8},
			{"4c", 9},
			{"4d", 10},
			{"4h", 11},
			{"4s", 12},
			{"5c", 13},
			{"5d", 14},
			{"5h", 15},
			{"5s", 16},
			{"6c", 17},
			{"6d", 18},
			{"6h", 19},
			{"6s", 20},
			{"7c", 21},
			{"7d", 22},
			{"7h", 23},
			{"7s", 24},
			{"8c", 25},
			{"8d", 26},
			{"8h", 27},
			{"8s", 28},
			{"9c", 29},
			{"9d", 30},
			{"9h", 31},
			{"9s", 32},
			{"tc", 33},
			{"td", 34},
			{"th", 35},
			{"ts", 36},
			{"jc", 37},
			{"jd", 38},
			{"jh", 39},
			{"js", 40},
			{"qc", 41},
			{"qd", 42},
			{"qh", 43},
			{"qs", 44},
			{"kc", 45},
			{"kd", 46},
			{"kh", 47},
			{"ks", 48},
			{"ac", 49},
			{"ad", 50},
			{"ah", 51},
			{"as", 52},
		},
	}

	for _, pair := range deck.stringNumberPair {
		deck.numberToString[pair.number] = pair.str
		deck.stringToNumber[pair.str] = pair.number
	}

	return deck
}

func (d *Deck) SameSuit(c []uint8) bool {

	lastSuit := ""

	cards, err := d.CardNumbersToStrings(c)

	if err != nil {
		return false
	}

	for _, c := range cards {

		current := c[1:]

		switch {
		case lastSuit == current:
		case lastSuit == "":
			lastSuit = current
		default:
			return false
		}
	}

	return true
}

func (d *Deck) NumberToString(c uint8) string {

	if s, ok := d.numberToString[c]; ok {
		return s
	}

	return "invalid"
}

func (d *Deck) CardNumbersToStrings(cards []uint8) ([]string, error) {

	mapped := make([]string, len(cards))

	for i, n := range cards {

		string, ok := d.numberToString[n]
		if !ok {
			return nil, fmt.Errorf("%d is not a valid card", n)
		}
		mapped[i] = string
	}

	return mapped, nil
}

func (d *Deck) CardStringsToNumbers(cards []string) ([]uint8, error) {

	mapped := make([]uint8, len(cards))

	for i, s := range cards {

		number, ok := d.stringToNumber[strings.ToLower(s)]
		if !ok {
			return nil, errors.New(s + " is not a valid card")
		}
		mapped[i] = number
	}

	return mapped, nil
}

func (d *Deck) AllNumberValues() []uint8 {

	if len(d.stringToNumber) != 52 {
		panic("deck not complete")
	}

	result := make([]uint8, 0, len(d.stringToNumber))

	for _, v := range d.stringNumberPair {
		result = append(result, v.number)
	}

	return result
}

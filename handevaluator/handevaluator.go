package handevaluator

import (
	"encoding/binary"
	"os"
)

type EvaluatedHand struct {
	HandName string
	HandRank uint32
	Value    uint32
}

type HandEvaluator struct {
	buffer []byte
}

const InvalidHand string = "invalid hand"

var handTypes = []string{
	InvalidHand,
	"high card",
	"one pair",
	"two pairs",
	"three of a kind",
	"straight",
	"flush",
	"full house",
	"four of a kind",
	"straight flush",
}

func New() (HandEvaluator, error) {
	h := HandEvaluator{}
	err := h.intialize()

	if err != nil {
		return h, err
	}
	return h, nil
}

func (e *HandEvaluator) intialize() error {
	file, err := os.Open("HandRanks.dat")
	if err != nil {
		return err
	}

	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		return err
	}

	filesize := fileinfo.Size()
	e.buffer = make([]byte, filesize)

	_, err = file.Read(e.buffer)
	if err != nil {
		return err
	}
	return nil
}

func (e *HandEvaluator) Eval(hand []int) EvaluatedHand {
	p := uint32(53)

	for _, c := range hand {
		start := 4 * (p + uint32(c))
		p = binary.LittleEndian.Uint32(e.buffer[start : start+4])
	}

	return EvaluatedHand{
		Value:    p,
		HandRank: p & 0x00000fff,
		HandName: handTypes[p>>12],
	}
}

func HandTypes() []string {
	return handTypes
}

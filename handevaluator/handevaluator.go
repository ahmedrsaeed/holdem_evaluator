package handevaluator

import (
	"encoding/binary"
	"os"
)

type EvaluatedHand struct {
	HandName string
	//HandRank uint32
	Value uint32
}

type HandEvaluator struct {
	buffer    []byte
	handTypes []string
}

const InvalidHand string = "invalid hand"

func HandTypes() []string {
	return []string{
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
}

func New() (HandEvaluator, error) {
	h := HandEvaluator{}
	h.handTypes = HandTypes()
	err := h.intializeBuffer()

	if err != nil {
		return h, err
	}
	return h, nil
}

func (e *HandEvaluator) intializeBuffer() error {
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

func (e *HandEvaluator) fromBuffer(p uint32, hand ...[]int) uint32 {

	for _, subset := range hand {
		for _, c := range subset {
			start := 4 * (p + uint32(c))
			p = binary.LittleEndian.Uint32(e.buffer[start : start+4])
		}
	}

	return p
}

func (e *HandEvaluator) Eval(partial ...[]int) func([]int) EvaluatedHand {

	partialResult := e.fromBuffer(uint32(53), partial...)

	return func(rest []int) EvaluatedHand {

		finalResult := e.fromBuffer(partialResult, rest)
		return EvaluatedHand{
			Value: finalResult,
			//HandRank: finalP & 0x00000fff,
			HandName: e.handTypes[finalResult>>12],
		}
	}
}

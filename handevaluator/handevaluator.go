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
	ranks []uint32
	//handTypes []string
}

const InvalidHandIndex = 0

func HandTypes() []string {
	return []string{
		"invalid hand",
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
	//h.handTypes = HandTypes()
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
	buffer := make([]byte, filesize)
	e.ranks = make([]uint32, filesize/4)

	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	for i := 0; i < int(filesize); i += 4 {
		e.ranks[i/4] = binary.LittleEndian.Uint32(buffer[i:])
	}
	return nil
}

func (e *PartialEvaluation) fromBuffer(p uint32, c uint8) uint32 {
	return e.ranks[p+uint32(c)]

}

func (e *HandEvaluator) PartialEvaluation(partial ...[]uint8) PartialEvaluation {

	partialEvaluation := PartialEvaluation{ranks: e.ranks}
	var partialResult uint32 = 53
	for _, subset := range partial {
		for _, c := range subset {
			partialResult = partialEvaluation.fromBuffer(partialResult, c)
		}
	}

	partialEvaluation.partial = partialResult
	return partialEvaluation
}

type PartialEvaluation struct {
	partial uint32
	ranks   []uint32
}

func (e *PartialEvaluation) Eval(a uint8, b uint8) (uint32, uint32) {

	finalResult := e.fromBuffer(e.fromBuffer(e.partial, a), b)

	return finalResult, finalResult >> 12
}

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

func (e *HandEvaluator) fromBuffer(p uint32, c uint8) uint32 {
	start := 4 * (p + uint32(c))
	return binary.LittleEndian.Uint32(e.buffer[start : start+4])

}

func (e *HandEvaluator) CreateFrom(partial ...[]uint8) func(uint8, uint8) EvaluatedHand {

	var partialResult uint32 = 53
	for _, subset := range partial {
		for _, c := range subset {
			partialResult = e.fromBuffer(partialResult, c)
		}
	}

	return func(a uint8, b uint8) EvaluatedHand {

		finalResult := e.fromBuffer(e.fromBuffer(partialResult, a), b)

		return EvaluatedHand{
			Value: finalResult,
			//HandRank: finalP & 0x00000fff,
			HandName: e.handTypes[finalResult>>12],
		}
	}
}

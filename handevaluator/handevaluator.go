package handevaluator

import (
	"encoding/binary"
	"fmt"
	"os"
)

type HandEvaluator struct {
	buffer []byte
}

func New() HandEvaluator {
	h := HandEvaluator{}
	h.intialize()
	return h
}

func (e *HandEvaluator) intialize() {
	file, err := os.Open("HandRanks.dat")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	filesize := fileinfo.Size()
	e.buffer = make([]byte, filesize)

	bytesread, err := file.Read(e.buffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("bytes read: ", bytesread)

}

func (e *HandEvaluator) Eval(hand []int) uint32 {
	p := uint32(53)

	for _, c := range hand {
		start := 4 * (p + uint32(c))
		p = binary.LittleEndian.Uint32(e.buffer[start : start+4])
	}

	return p
}

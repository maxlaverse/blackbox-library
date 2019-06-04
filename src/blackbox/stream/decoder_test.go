package stream

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadByteOnce(t *testing.T) {
	r := bytes.NewReader([]byte{1, 2, 3})
	decoder := NewDecoder(r)

	val, err := decoder.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, uint8(1), val)
}

func TestReadByteTwice(t *testing.T) {
	r := bytes.NewReader([]byte{1, 2, 3})
	decoder := NewDecoder(r)

	val, err := decoder.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, uint8(1), val)

	val, err = decoder.ReadByte()

	assert.Nil(t, err)
	assert.Equal(t, uint8(2), val)
}

func TestNextByteTwice(t *testing.T) {
	r := bytes.NewReader([]byte{1, 2, 3})
	decoder := NewDecoder(r)

	val, err := decoder.NextByte()
	assert.Nil(t, err)
	assert.Equal(t, uint8(1), val)

	val, err = decoder.NextByte()
	assert.Nil(t, err)
	assert.Equal(t, uint8(1), val)
}

func TestEOFNotReached(t *testing.T) {
	r := bytes.NewReader([]byte{1, 2})
	decoder := NewDecoder(r)

	_, err := decoder.ReadByte()
	assert.Nil(t, err)

	eof, err := decoder.EOF()
	assert.Nil(t, err)
	assert.False(t, eof)
}

func TestEOFReached(t *testing.T) {
	r := bytes.NewReader([]byte{1, 2})
	decoder := NewDecoder(r)

	_, err := decoder.ReadByte()
	assert.Nil(t, err)

	_, err = decoder.ReadByte()
	assert.Nil(t, err)

	eof, err := decoder.EOF()
	assert.Nil(t, err)
	assert.True(t, eof)
}

func TestReadUnsignedVB(t *testing.T) {
	inputArray := [][]byte{
		[]byte{55, 99},
		[]byte{179, 12, 99},
		[]byte{189, 254, 13, 99},
	}
	outputArray := []uint32{0x37, 0x633, 0x37f3d}

	for testIndex, input := range inputArray {
		t.Run(fmt.Sprintf("for %v", input), func(t *testing.T) {
			r := bytes.NewReader(input)
			decoder := NewDecoder(r)

			val, err := decoder.ReadUnsignedVB()
			assert.Nil(t, err)
			assert.Equal(t, outputArray[testIndex], val)
		})

	}
}

func TestReadSignedVB(t *testing.T) {
	inputArray := [][]byte{
		[]byte{55, 99},
		[]byte{179, 12, 99},
		[]byte{189, 254, 13, 99},
	}
	outputArray := []int32{-28, -794, -114591}

	for testIndex, input := range inputArray {
		t.Run(fmt.Sprintf("for %v", input), func(t *testing.T) {
			r := bytes.NewReader(input)
			decoder := NewDecoder(r)

			val, err := decoder.ReadSignedVB()
			assert.Nil(t, err)
			assert.Equal(t, outputArray[testIndex], val)
		})

	}
}

func TestReadTag8_8SVB(t *testing.T) {
	inputArray := [][]byte{
		[]byte{51, 15, 57, 112, 99},
		[]byte{3, 12, 148, 99},
		[]byte{189, 254, 13, 99, 1, 2, 3, 4, 99, 15},
	}
	outputArray := [][]int32{
		[]int32{-8, -29, 0, 0, 56, -50, 0, 0},
		[]int32{6, 6346, 0, 0, 0, 0, 0, 0},
		[]int32{895, 0, -50, -1, 1, -2, 0, 2},
	}

	for testIndex, input := range inputArray {
		t.Run(fmt.Sprintf("for %v", input), func(t *testing.T) {
			r := bytes.NewReader(input)
			decoder := NewDecoder(r)

			val, err := decoder.ReadTag8_8SVB(3)
			assert.Nil(t, err)
			assert.Equal(t, outputArray[testIndex], val)
		})
	}
}

func TestReadTag8_8SVB2(t *testing.T) {
	inputArray := [][]byte{
		[]byte{51, 15, 57, 112, 99},
		[]byte{3, 12, 148, 99},
		[]byte{189, 254, 13, 99, 1, 2, 3, 4, 99, 15},
	}
	outputArray := [][]int32{
		[]int32{-26, 0, 0, 0, 0, 0, 0, 0},
		[]int32{-2, 0, 0, 0, 0, 0, 0, 0},
		[]int32{-114591, 0, 0, 0, 0, 0, 0, 0},
	}

	for testIndex, input := range inputArray {
		t.Run(fmt.Sprintf("for %v", input), func(t *testing.T) {
			r := bytes.NewReader(input)
			decoder := NewDecoder(r)

			val, err := decoder.ReadTag8_8SVB(1)
			assert.Nil(t, err)
			assert.Equal(t, outputArray[testIndex], val)
		})
	}
}

func TestReadTag2_3S32(t *testing.T) {
	// first byte = leadByte leadByte sizeByte3 sizeByte3 sizeByte2 sizeByte2 sizeByte1 sizeByte1
	inputArray := [][]byte{
		// 2 bits fields - lead byte 30 (00 01 11 10)
		[]byte{30},

		// 6 bits fields - lead byte 156 (10 011100), 15 (00 001111), 57 (00 111001)
		[]byte{156, 15, 57},

		// 2 bits fields - lead byte 63 (00 11 11 11)
		[]byte{63},

		// 4 bits fields - lead byte 127 (01 11 1111), 84 (0101 0100)
		[]byte{127, 84},

		// 6 bits fields - lead byte 189 (10 111101), 254 (11 111110), 13 (00 001101)
		[]byte{189, 254, 13},

		// variable bits fields - lead byte 192 (11 00 00 00)15, 24, 77
		[]byte{192, 15, 24, 77},

		// variable bits fields - lead byte 199 (11 00 01 11), 65 (01000001), 77 (01001101), 155
		[]byte{199, 1, 1, 1, 1, 65, 77, 155},
	}
	outputArray := [][]int32{
		// 1 (000000 01), -1 (111111 11), -2 (111111 10)
		[]int32{1, -1, -2, 0, 0, 0, 0, 0},

		// 28 (00 011100), 15 (00 001111), -7 (11 111001)
		[]int32{28, 15, -7, 0, 0, 0, 0, 0},

		// -1 (11), -1 (11), -1 (11)
		[]int32{-1, -1, -1, 0, 0, 0, 0, 0},

		// -1 (1111), 5 (0101), 4 (0100)
		[]int32{-1, 5, 4, 0, 0, 0, 0, 0},

		// -3 (111101), -2 (111110), 13 (001101)
		[]int32{-3, -2, 13, 0, 0, 0, 0, 0},

		[]int32{15, 24, 77, 0, 0, 0, 0, 0},

		// 16843009 (00000001 0000000 10000000 100000001), 19777 (01001101 01000001), 155
		[]int32{16843009, 19777, 155, 0, 0, 0, 0, 0},
	}

	for testIndex, input := range inputArray {
		t.Run(fmt.Sprintf("for %v", input), func(t *testing.T) {
			r := bytes.NewReader(input)
			decoder := NewDecoder(r)

			val, err := decoder.ReadTag2_3S32()
			assert.Nil(t, err)
			assert.Equal(t, outputArray[testIndex], val)
		})
	}
}

func TestReadTag8_4S16V2(t *testing.T) {
	inputArray := [][]byte{
		[]byte{156, 15, 57, 112, 99},
		[]byte{63, 15, 57, 112, 99, 1, 2, 3, 4, 5, 6, 91, 2, 3, 4, 5, 6, 9},
		[]byte{127, 84, 1, 56, 99, 1, 2, 3, 4, 5, 6, 91, 2, 3, 4, 5, 6, 9},
		[]byte{189, 254, 13, 99, 1, 2, 3, 4, 99, 15},
	}
	outputArray := [][]int32{
		[]int32{0, 57, 7, 6, 0, 0, 0, 0},
		[]int32{57, 99, 2, 0, 0, 0, 0, 0},
		[]int32{1, 99, 2, 0, 0, 0, 0, 0},
		[]int32{-1, 214, 16, 32, 0, 0, 0, 0},
	}

	for testIndex, input := range inputArray {
		t.Run(fmt.Sprintf("for %v", input), func(t *testing.T) {
			r := bytes.NewReader(input)
			decoder := NewDecoder(r)

			val, err := decoder.ReadTag8_4S16V2()
			assert.Nil(t, err)
			assert.Equal(t, outputArray[testIndex], val)
		})
	}
}

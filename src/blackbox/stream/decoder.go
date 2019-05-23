package stream

import (
	"bufio"
	"io"
)

/*
* See https://cleanflight.readthedocs.io/en/stable/development/Blackbox%20Internals/ for more information about types end predictors.
 */
const (
	fieldZero  = 0
	field4Bit  = 1
	field8Bit  = 2
	field16Bit = 3
)

// Decoder can read various form of encoded bytes
type Decoder struct {
	reader        *bufio.Reader
	statBytesRead int64
}

// NewDecoder returns a new instance of a Decoder
func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{
		reader:        bufio.NewReaderSize(reader, 8*1024),
		statBytesRead: 0,
	}
}

// BytesRead returns the number of bytes read
func (d *Decoder) BytesRead() int64 {
	return d.statBytesRead
}

// ReadByte reads one byte
func (d *Decoder) ReadByte() (byte, error) {
	bytes, err := d.ReadBytes(1)
	if err != nil {
		return 0, err
	}

	return bytes[0], nil
}

// ReadInt reads one byte as integer
func (d *Decoder) ReadInt() (int32, error) {
	bytes, err := d.ReadBytes(1)
	if err != nil {
		return 0, err
	}

	return int32(bytes[0]), nil
}

// ReadBytes reads multiple bytes
func (d *Decoder) ReadBytes(number int) ([]byte, error) {
	bytes := make([]byte, number)
	n, err := d.reader.Read(bytes)
	if err != nil {
		return nil, errorWithStack(err)
	}
	d.statBytesRead += int64(n)
	return bytes, nil
}

// NextByte returns the next byte without changing the file pointer
func (d *Decoder) NextByte() (byte, error) {
	bytes, err := d.reader.Peek(1)
	if err != nil {
		return 0, errorWithStack(err)
	}
	return bytes[0], nil
}

// EOF returns true if the end of file was reached
func (d *Decoder) EOF() (bool, error) {
	_, err := d.NextByte()
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// ReadUnsignedVB reads the most straightforward encoding. This encoding uses
// the lower 7 bits of an encoded byte to store the lower 7 bits of the field's
// value. The high bit of that encoded byte is set to one if more than 7 bits
// are required to store the value. If the value did exceed 7 bits, the lower 7
// bits of the value (which were written to the log) are removed from the
// value (by right shift), and the encoding process begins again with the new
// value.
func (d *Decoder) ReadUnsignedVB() (uint32, error) {
	result := uint32(0)
	i := uint(0)
	for {
		value, err := d.ReadByte()
		if err != nil {
			return result, err
		}

		result = result + (uint32(value)&0x7F)<<(7*i)

		if value < 128 {
			break
		}
		i = i + 1
	}
	return result, nil
}

// ReadSignedVB applies a pre-processing step to fold negative values into
// positive ones, then the resulting unsigned number is encoded using unsigned
// variable byte encoding.
func (d *Decoder) ReadSignedVB() (int32, error) {
	val, err := d.ReadUnsignedVB()
	if err != nil {
		return 0, err
	}
	return zigzagDecode(val), nil
}

// ReadTag8_8SVB first an 8-bit (one byte) header is written. This header
// has its bits set to zero when the corresponding field (from a maximum of 8
// fields) is set to zero, otherwise the bit is set to one. The
// least-signficant bit in the header corresponds to the first field to be
// written. This header is followed by the values of only the fields which are
// non-zero, written using signed variable byte encoding.
func (d *Decoder) ReadTag8_8SVB(valueCount int) ([]int32, error) {
	values := make([]int32, 8)
	if valueCount == 1 {
		val, err := d.ReadSignedVB()
		if err != nil {
			return values, err
		}
		values[0] = val
	} else {
		val, err := d.ReadByte()
		if err != nil {
			return values, err
		}
		header := uint8(val)
		for i := 0; i < 8; i++ {
			if header&0x01 == 0x01 {
				val, err := d.ReadSignedVB()
				if err != nil {
					return values, err
				}
				values[i] = val
			} else {
				values[i] = 0
			}
			header = header >> 1
		}
	}
	return values, nil
}

// ReadTag2_3S32 returns
func (d *Decoder) ReadTag2_3S32() ([]int32, error) {
	values := make([]int32, 8)
	leadByte, err := d.ReadByte()
	if err != nil {
		return values, err
	}

	switch leadByte >> 6 {
	case 0:
		// 2-bit fields
		values[0] = SignExtend2Bit(uint8((leadByte >> 4) & 0x03))
		values[1] = SignExtend2Bit(uint8((leadByte >> 2) & 0x03))
		values[2] = SignExtend2Bit(uint8(leadByte & 0x03))
	case 1:
		// 4-bit fields
		values[0] = SignExtend4Bit(uint8(leadByte & 0x0F))

		leadByte, err = d.ReadByte()
		if err != nil {
			return values, err
		}
		values[1] = SignExtend4Bit(uint8(leadByte >> 4))
		values[2] = SignExtend4Bit(uint8(leadByte & 0x0F))
	case 2:
		// 6-bit fields
		values[0] = SignExtend6Bit(uint8(leadByte & 0x3F))

		leadByte, err := d.ReadByte()
		if err != nil {
			return values, err
		}
		values[1] = SignExtend6Bit(uint8(leadByte & 0x3F))

		leadByte, err = d.ReadByte()
		if err != nil {
			return values, err
		}
		values[2] = SignExtend6Bit(uint8(leadByte & 0x3F))
	case 3:
		// Fields are 8, 16 or 24 bits, read selector to figure out which field is which size
		for i := 0; i < 3; i++ {
			switch leadByte & 0x03 {
			case 0: // 8-bit
				byte1, err := d.ReadInt()
				if err != nil {
					return values, err
				}
				values[i] = byte1
			case 1: // 16-bit
				byte1, err := d.ReadInt()
				if err != nil {
					return values, err
				}
				byte2, err := d.ReadInt()
				if err != nil {
					return values, err
				}

				values[i] = byte1 | byte2<<8
			case 2: // 24-bit
				byte1, err := d.ReadInt()
				if err != nil {
					return values, err
				}
				byte2, err := d.ReadInt()
				if err != nil {
					return values, err
				}
				byte3, err := d.ReadInt()
				if err != nil {
					return values, err
				}

				values[i] = SignExtend24Bit(uint32(byte1 | (byte2 << 8) | (byte3 << 16)))
			case 3: // 32-bit
				byte1, err := d.ReadInt()
				if err != nil {
					return values, err
				}
				byte2, err := d.ReadInt()
				if err != nil {
					return values, err
				}
				byte3, err := d.ReadInt()
				if err != nil {
					return values, err
				}
				byte4, err := d.ReadInt()
				if err != nil {
					return values, err
				}
				// Sign-extend
				values[i] = (byte1 | (byte2 << 8) | (byte3 << 16) | (byte4 << 24))
			}
			leadByte >>= 2
		}
	}
	return values, nil
}

//ReadTag8_4S16V1 returns
func (d *Decoder) ReadTag8_4S16V1() ([]int32, error) {
	// TODO: Implement
	panic("Not implemented")
}

// ReadTag8_4S16V2 returns
func (d *Decoder) ReadTag8_4S16V2() ([]int32, error) {
	values := make([]int32, 8)

	selector, err := d.ReadByte()
	if err != nil {
		return values, err
	}
	buffer := uint8(0)
	char1 := uint8(0)
	char2 := uint8(0)

	//Read the 4 values from the stream
	nibbleIndex := 0
	for i := 0; i < 4; i++ {
		switch selector & 0x03 {
		case fieldZero:
			values[i] = 0

		case field4Bit: // Two 4-bit fields
			if nibbleIndex == 0 {
				val, err := d.ReadByte()
				if err != nil {
					return values, err
				}
				buffer = uint8(val)
				values[i] = SignExtend4Bit(buffer >> 4)
				nibbleIndex = 1
			} else {
				values[i] = SignExtend4Bit(buffer & 0x0F)
				nibbleIndex = 0
			}

		case field8Bit: // 8-bit field
			if nibbleIndex == 0 {
				//Sign extend...
				val, err := d.ReadByte()
				if err != nil {
					return values, err
				}

				values[i] = int32(val)
			} else {
				char1 = buffer << 4
				val, err := d.ReadByte()
				if err != nil {
					return values, err
				}
				buffer = uint8(val)

				char1 |= buffer >> 4
				values[i] = int32(char1)
			}

		case field16Bit: // 16-bit field
			if nibbleIndex == 0 {
				val, err := d.ReadByte()
				if err != nil {
					return values, err
				}
				char1 = uint8(val)
				val, err = d.ReadByte()
				if err != nil {
					return values, err
				}
				char2 = uint8(val)

				//Sign extend...
				values[i] = int32(uint16(((char1 << 8) | char2)))
			} else {
				/*
				 * We're in the low 4 bits of the current buffer, then one byte, then the high 4 bits of the next
				 * buffer.
				 */
				val, err := d.ReadByte()
				if err != nil {
					return values, err
				}
				char1 = uint8(val)
				val, err = d.ReadByte()
				if err != nil {
					return values, err
				}
				char2 = uint8(val)

				values[i] = int32(uint16(((buffer << 12) | (char1 << 4) | (char2 >> 4))))

				buffer = char2
			}

		}

		selector >>= 2

	}
	return values, nil
}

func zigzagDecode(value uint32) int32 {
	return int32((value >> 1) ^ -(value & 1))
}

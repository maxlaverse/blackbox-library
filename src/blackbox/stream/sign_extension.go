package stream

// SignExtend2Bit signs
func SignExtend2Bit(val uint8) int64 {
	if val&0x02 == 0x02 {
		return int64(int8(val | 0xFC))
	}
	return int64(val)
}

// SignExtend4Bit signs
func SignExtend4Bit(nibble uint8) int64 {
	if nibble&0x08 == 0x08 {
		return int64(int8(nibble | 0xF0))
	}
	return int64(nibble)
}

// SignExtend6Bit signs
func SignExtend6Bit(nibble uint8) int64 {
	if nibble&0x20 == 0x20 {
		return int64(int8(nibble | 0xC0))
	}
	return int64(nibble)
}

// SignExtend14Bit signs
func SignExtend14Bit(val uint16) int64 {
	if val&0x2000 == 0x2000 {
		return int64(int16(val | 0xC000))
	}
	return int64(val)
}

// SignExtend24Bit signs
func SignExtend24Bit(val uint32) int64 {
	if val&0x800000 == 0x800000 {
		return int64(int32(val | 0xFF000000))
	}
	return int64(val)
}

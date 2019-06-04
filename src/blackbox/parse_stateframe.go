package blackbox

import (
	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
	"github.com/pkg/errors"
)

const (
	// EncodingSignedVB is for signed variable-byte
	EncodingSignedVB = 0

	// EncodingUnsignedVB is for Unsigned variable-byte
	EncodingUnsignedVB = 1

	// EncodingNeg14Bits is for Unsigned variable-byte but we negate the value before storing value is 14 bits
	EncodingNeg14Bits = 3

	// EncodingTag8_8SVB is for First, an 8-bit (one byte) header is written. This header has its bits set to zero when the corresponding field (from a maximum of 8 fields) is set to zero, otherwise the bit is set to one.
	EncodingTag8_8SVB = 6

	// EncodingTag2_3S32 is for a 2-bit header is written, followed by 3 signed field values of up to 32 bits each.
	EncodingTag2_3S32 = 7

	// EncodingTag8_4S16 is for an 8-bit header is written, followed by 4 signed field values of up to 16 bits each.
	EncodingTag8_4S16 = 8

	// EncodingNull is for nothing is written to the file take value to be zero
	EncodingNull = 9
)

// parseStateFrame parse any kind of data frame and applies prediction
func parseStateFrame(frameDef LogDefinition, fields []FieldDefinition, previousFrame, previousPreviousFrame *MainFrame, dec *stream.Decoder, disablePredicator bool, skippedFrames int32) ([]int32, error) {
	frameValues := make([]int32, len(fields))
	framesToSkip := 0
	for i, field := range fields {
		// Skip frames we exceptionnaly already read in the previous round
		if framesToSkip > 0 {
			framesToSkip--
			continue
		}

		// Don't calculate absolute values based on previous Frames
		if disablePredicator && field.Name != FieldIteration {
			field.Predictor = PredictorZero
		}

		// Simple predicator that increments fields. No need to do more
		if field.Predictor == PredictorInc {
			frameValues[i] = skippedFrames + 1
			if previousFrame != nil {
				frameValues[i] += previousFrame.values[i]
			}
			continue
		}

		// Decode the value for the current field
		// This might lead to the next adjacent fields to be processed as well when reading arrays
		value := int32(0)
		switch field.Encoding {
		case EncodingSignedVB:
			val, err := dec.ReadSignedVB()
			if err != nil {
				return nil, err
			}

			value = val
		case EncodingUnsignedVB:
			val, err := dec.ReadUnsignedVB()
			if err != nil {
				return nil, err
			}

			value = int32(val)
		case EncodingNeg14Bits:
			val, err := dec.ReadUnsignedVB()
			if err != nil {
				return nil, err
			}

			value = -int32(stream.SignExtend14Bit(uint16(val)))
		case EncodingTag8_8SVB:
			vals, err := dec.ReadTag8_8SVB(field.GroupCount)
			if err != nil {
				return nil, err
			}
			for j := 0; j < field.GroupCount; j++ {
				v, err := ApplyPrediction(frameDef, vals, i+j, int(field.Predictor), vals[j], previousFrame, previousPreviousFrame)
				if err != nil {
					return nil, err
				}
				frameValues[i+j] = v
			}
			framesToSkip = field.GroupCount - 1
			continue
		case EncodingTag2_3S32:
			vals, err := dec.ReadTag2_3S32()
			if err != nil {
				return nil, err
			}
			for j := 0; j < 3; j++ {
				v, err := ApplyPrediction(frameDef, vals, i+j, int(field.Predictor), vals[j], previousFrame, previousPreviousFrame)
				if err != nil {
					return nil, err
				}
				frameValues[i+j] = v
			}
			framesToSkip = 2
			continue
		case EncodingTag8_4S16:
			var vals []int32
			var err error
			if frameDef.DataVersion == 1 {
				vals, err = dec.ReadTag8_4S16V1()
			} else {
				vals, err = dec.ReadTag8_4S16V2()
			}
			if err != nil {
				return nil, err
			}
			for j := 0; j < 4; j++ {
				v, err := ApplyPrediction(frameDef, vals, i+j, int(field.Predictor), vals[j], previousFrame, previousPreviousFrame)
				if err != nil {
					return nil, err
				}
				frameValues[i+j] = v
			}
			framesToSkip = 3
			continue
		case EncodingNull:
			value = 0
		default:
			return nil, errors.Errorf("Unsupported decoding '%d' for field '%s'", field.Encoding, field.Name)
		}

		v, err := ApplyPrediction(frameDef, frameValues, i, int(field.Predictor), value, previousFrame, previousPreviousFrame)
		if err != nil {
			return nil, err
		}
		frameValues[i] = v
	}
	return frameValues, nil
}

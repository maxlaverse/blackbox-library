package blackbox

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
	"github.com/stretchr/testify/assert"
)

var encodedFrameI = []byte{73, 128, 163, 3, 251, 171, 176, 26, 0, 4, 0, 6, 3, 3, 10, 6, 0, 0, 0, 3, 0, 0, 192, 9, 1, 0, 0, 176, 3, 233, 1, 166, 11, 145, 6, 1, 3, 0, 54, 116, 250, 34, 4, 14, 0, 0, 140, 3, 35, 36, 38}

var encodedFrameS = []byte{83, 129, 128, 32, 8, 5, 80, 230, 7, 2, 0, 2, 0, 7, 9, 0, 0, 0, 0, 0, 0, 0, 2, 1, 0, 3, 1, 4, 10, 3, 0, 9, 32, 27, 8, 80, 2, 0, 2, 1, 0, 1, 8, 0, 0, 0, 0, 0, 0, 1, 0, 2, 0, 1, 0, 6, 18, 0, 0, 14, 0, 0, 11, 80, 8, 4, 6, 0, 0, 8, 24, 0, 0, 0, 0, 0, 0, 5, 3, 2, 0, 3, 3, 2, 18, 2, 0, 48, 91, 86, 39, 80, 11, 1, 6, 0, 0, 14, 40, 0, 0, 0, 0, 0, 0, 0, 7, 0, 0, 1, 1, 2, 18, 4, 0, 84, 157, 1, 160, 1, 81, 80, 6, 0, 4, 2, 0, 4, 42, 0, 0, 0, 0, 0, 3, 168, 2, 141, 1, 0, 5, 1, 2, 0, 0, 4, 14, 4, 0, 112, 147, 1, 152, 1, 115, 80, 2, 0, 0, 0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 3, 0, 2, 0, 0, 2, 4, 4, 0, 86, 91, 94, 87, 80, 13, 2, 0, 0, 0, 0, 8, 0, 0, 0, 1, 16, 0, 0, 1, 0, 0, 0, 4, 8, 3, 11, 2, 0, 32, 39, 42, 29, 80, 14, 1, 3, 0, 0, 0, 13, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 2, 4, 3, 21}

var encodedFramesE = [][]byte{
	[]byte{69, 14, 128, 158, 3, 248, 201, 166, 26, 73, 128, 158, 3, 248, 201, 166, 26, 1, 7, 1, 10, 3, 1, 1, 59, 0, 0, 0, 1, 0, 8, 192, 9, 1, 0, 4, 176, 3, 130, 2, 230, 11, 145, 6, 0, 6, 6, 120, 16, 240, 34, 9, 27, 2, 0, 205, 2, 130, 2, 3, 160, 2, 69, 0, 145, 139, 246, 19, 69, 30, 129, 128, 32, 1, 83, 129, 128, 32, 8, 5, 80, 230, 7, 2, 0, 2, 0, 7, 9, 0, 0, 0, 0, 0, 0, 0, 2, 1, 0, 3, 1, 4, 10, 3, 0, 9, 32, 27, 8, 80, 2, 0, 2, 1, 0, 1, 8, 0, 0, 0, 0, 0, 0, 1, 0, 2, 0, 1, 0, 6, 18, 0, 0, 14, 0, 0, 11, 80, 8, 4, 6, 0, 0, 8, 24, 0, 0, 0, 0, 0, 0, 5, 3, 2, 0, 3, 3, 2, 18, 2, 0, 48, 91, 86, 39, 80, 11, 1, 6, 0, 0, 14, 40, 0, 0, 0, 0, 0, 0, 0, 7, 0, 0, 1, 1, 2, 18, 4, 0, 84, 157, 1, 160, 1, 81, 80, 6, 0, 4, 2, 0, 4, 42, 0, 0, 0, 0, 0, 3, 168, 2, 141, 1, 0, 5, 1, 2, 0, 0, 4, 14, 4, 0, 112, 147, 1, 152, 1, 115, 80, 2, 0, 0},
	[]byte{69, 0, 145, 139, 246, 19, 69, 30, 129, 128, 32, 1, 83, 129, 128, 32, 8, 5, 80, 230, 7, 2, 0, 2, 0, 7, 9, 0, 0, 0, 0, 0, 0, 0, 2, 1, 0, 3, 1, 4, 10, 3, 0, 9, 32, 27, 8, 80, 2, 0, 2, 1, 0, 1, 8, 0, 0, 0, 0, 0, 0, 1, 0, 2, 0, 1, 0, 6, 18, 0, 0, 14, 0, 0, 11, 80, 8, 4, 6, 0, 0, 8, 24, 0, 0, 0, 0, 0, 0, 5, 3, 2, 0, 3, 3, 2, 18, 2, 0, 48, 91, 86, 39, 80, 11, 1, 6, 0, 0, 14, 40, 0, 0, 0, 0, 0, 0, 0, 7, 0, 0, 1, 1, 2, 18, 4, 0, 84, 157, 1, 160, 1, 81, 80, 6, 0, 4, 2, 0, 4, 42, 0, 0, 0, 0, 0, 3, 168, 2, 141, 1, 0, 5, 1, 2, 0, 0, 4, 14, 4, 0, 112, 147, 1, 152, 1, 115, 80, 2, 0, 0, 0, 0, 0, 24, 0, 0, 0, 0, 0, 0, 0, 3, 0, 2, 0, 0, 2, 4, 4, 0, 86, 91, 94, 87, 80, 13, 2, 0, 0, 0, 0, 8, 0, 0, 0, 1, 16, 0, 0, 1, 0, 0, 0, 4, 8, 3, 11, 2, 0, 32, 39, 42, 29, 80, 14, 1, 3, 0},
	[]byte{69, 255, 'E', 'n', 'd', ' ', 'o', 'f', ' ', 'l', 'o', 'g'},
}

var encodedFramesP = [][]byte{
	[]byte{80, 222, 7, 1, 0, 0, 0, 0, 28, 0, 0, 0, 0, 0, 0, 2, 0, 0, 1, 7, 7, 4, 12, 2, 0, 58, 51, 56, 61},
	[]byte{80, 24, 2, 6, 0, 0, 1, 20, 0, 0, 0, 0, 0, 0, 3, 3, 0, 0, 3, 3, 1, 6, 2, 0, 72, 79, 80, 67},
	[]byte{80, 27, 0, 4, 0, 0, 8, 18, 0, 0, 0, 0, 0, 0, 1, 3, 0, 0, 3, 3, 3, 7, 0, 0, 48, 75, 78, 47},
	[]byte{80, 20, 0, 3, 0, 0, 1, 2, 0, 0, 0, 0, 0, 0, 2, 0, 1, 0, 1, 1, 1, 13, 0, 0, 16, 11, 18, 17},
	[]byte{80, 5, 3, 9, 0, 0, 11, 25, 0, 0, 0, 0, 0, 3, 82, 71, 2, 8, 1, 0, 1, 0, 3, 17, 0, 0, 35, 98, 97, 38},
	[]byte{80, 6, 2, 0, 0, 0, 11, 47, 0, 0, 0, 0, 0, 0, 0, 4, 2, 0, 0, 0, 5, 19, 1, 0, 87, 156, 1, 155, 1, 90},
}

var decodedRawFrameI = []int32([]int32{53632, 55318011, 0, 2, 0, 3, -2, -2, 5, 3, 0, 0, 0, -2, 0, 0, 1216, -1, 0, 0, 216, -233, 723, 785, -1, -2, 0, 27, 58, 2237, 2, 7, 0, 0, 396, -18, 18, 19})

var decodedRawFrameP = []int32([]int32{1, 495, -1, 0, 0, 0, 0, 0, 0, 14, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, -1, -4, -4, 2, 6, 1, 0, 29, -26, 28, -31})

var decodedPredictedFrameI = []int32([]int32{53632, 55318011, 0, 2, 0, 3, -2, -2, 5, 3, 0, 0, 0, -2, 0, 0, 1216, -1, 0, 0, 216, 1632, 723, 785, -1, -2, 0, 27, 58, 2237, 2, 7, 0, 0, 584, 566, 602, 603})

var decodedPredictedFramesP = [][]int32{
	[]int32([]int32{53633, 55318506, -1, 2, 0, 3, -2, -2, 5, 17, 0, 0, 0, -2, 0, 0, 1216, -1, 0, 0, 216, 1632, 723, 785, 0, -2, 0, 26, 54, 2233, 4, 13, 1, 0, 613, 540, 630, 572}),
	[]int32([]int32{53634, 55319013, 0, 5, 0, 3, -2, -2, 4, 27, 0, 0, 0, -2, 0, 0, 1216, -1, 0, 0, 216, 1632, 723, 785, -2, -4, 0, 26, 54, 2233, 2, 13, 1, 0, 634, 513, 656, 553}),
	[]int32([]int32{53635, 55319506, 0, 7, 0, 3, -2, -2, 8, 36, 0, 0, 0, -2, 0, 0, 1216, -1, 0, 0, 216, 1632, 723, 785, -2, -5, 0, 26, 52, 2231, 1, 9, 1, 0, 647, 488, 682, 538}),
	[]int32([]int32{53636, 55320009, 0, 5, 0, 3, -2, -2, 7, 37, 0, 0, 0, -2, 0, 0, 1216, -1, 0, 0, 216, 1632, 723, 785, -1, -4, -1, 26, 52, 2231, 0, 4, 1, 0, 648, 494, 678, 536}),
	[]int32([]int32{53637, 55320509, -2, 0, 0, 3, -2, -2, 1, 24, 0, 0, 0, -2, 0, 0, 1216, -1, 0, 0, 216, 1673, 687, 785, 0, 0, -1, 26, 51, 2231, -2, -3, 1, 0, 629, 540, 631, 556}),
	[]int32([]int32{53638, 55321012, -1, 0, 0, 3, -2, -2, -5, 0, 0, 0, 0, -2, 0, 0, 1216, -1, 0, 0, 216, 1673, 687, 785, 0, 0, 0, 26, 51, 2231, -4, -10, 0, 0, 594, 595, 576, 591}),
}

func TestParseFrameIRaw(t *testing.T) {
	r := bytes.NewReader(encodedFrameI[1:])
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	res, err := parseStateFrame(frameDef, frameDef.FieldsI, nil, nil, dec, true, 0)
	assert.Nil(t, err)
	assert.Equal(t, decodedRawFrameI, res)
}

func TestParseFrameIPredicted(t *testing.T) {
	r := bytes.NewReader(encodedFrameI[1:])
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	res, err := parseStateFrame(frameDef, frameDef.FieldsI, nil, nil, dec, false, 0)

	assert.Nil(t, err)
	assert.Equal(t, decodedPredictedFrameI, res)
}

func TestParseFramePRaw(t *testing.T) {
	r := bytes.NewReader(encodedFramesP[0][1:])
	dec := stream.NewDecoder(r)
	frameDef := dummyFrameDefinition()

	res, err := parseStateFrame(frameDef, frameDef.FieldsP, nil, nil, dec, true, 0)

	assert.Nil(t, err)
	assert.Equal(t, decodedRawFrameP, res)
}

func TestParseFramePPredicted(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFrameI[1:], encodedFramesP[0][1:], encodedFramesP[1][1:], encodedFramesP[2][1:], encodedFramesP[3][1:], encodedFramesP[4][1:], encodedFramesP[5][1:]))
	dec := stream.NewDecoder(r)
	frameDef := dummyFrameDefinition()

	res, err := parseStateFrame(frameDef, frameDef.FieldsI, nil, nil, dec, false, 0)
	assert.Nil(t, err)
	assert.Equal(t, decodedPredictedFrameI, res)

	previousPreviousFrame := &MainFrame{
		values: res,
	}
	previousFrame := &MainFrame{
		values: res,
	}

	for idx, decodedFrame := range decodedPredictedFramesP {
		t.Run(fmt.Sprintf("for frame P%v", idx+1), func(t *testing.T) {
			res, err := parseStateFrame(frameDef, frameDef.FieldsP, previousFrame, previousPreviousFrame, dec, false, 0)
			assert.Nil(t, err)
			assert.Equal(t, decodedFrame, res)

			previousPreviousFrame = previousFrame
			previousFrame = &MainFrame{
				values: res,
			}
		})
	}
}

func buildStream(frame ...[]byte) []byte {
	frames := []byte{}

	for _, f := range frame {
		frames = append(frames, f...)
	}
	return frames
}

func dummyFrameDefinition() LogDefinition {
	frameDef := LogDefinition{
		Sysconfig: SysconfigType{
			MotorOutputLow:         188,
			MotorOutputHigh:        1850,
			RcRate:                 90,
			YawRate:                0,
			Acc1G:                  1,
			GyroScale:              1,
			Vbatscale:              110,
			Vbatmincellvoltage:     uint8(330 - 256),
			Vbatmaxcellvoltage:     uint8(440 - 256),
			Vbatwarningcellvoltage: uint8(350 - 256),
			CurrentMeterOffset:     0,
			CurrentMeterScale:      282,
			Vbatref:                1865,
			FirmwareType:           "Cleanflight",
			FrameIntervalI:         1,
			FrameIntervalPNum:      1,
			FrameIntervalPDenom:    1,
		},
		Headers: []Header{
			Header{
				Name:  "I interval",
				Value: "128",
			}, Header{
				Name:  "P interval",
				Value: "2",
			}, Header{
				Name:  "P ratio",
				Value: "64",
			},
		},
		FieldsI: []FieldDefinition{
			FieldDefinition{
				Name:     "loopIteration",
				Encoding: EncodingUnsignedVB,
			}, FieldDefinition{
				Name:     "time",
				Encoding: EncodingUnsignedVB,
			}, FieldDefinition{
				Name:     "axisP[0]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "axisP[1]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "axisP[2]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "axisI[0]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "axisI[1]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "axisI[2]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "axisD[0]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "axisD[1]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "axisF[0]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "axisF[1]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "axisF[2]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "rcCommand[0]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "rcCommand[1]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "rcCommand[2]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "rcCommand[3]",
				Encoding: EncodingUnsignedVB,
			}, FieldDefinition{
				Name:     "setpoint[0]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "setpoint[1]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "setpoint[2]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "setpoint[3]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:      "vbatLatest",
				Encoding:  EncodingNeg14Bits,
				Predictor: PredictorVbatRef,
			}, FieldDefinition{
				Name:     "amperageLatest",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "rssi",
				Encoding: EncodingUnsignedVB,
			}, FieldDefinition{
				Name:     "gyroADC[0]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "gyroADC[1]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "gyroADC[2]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "accSmooth[0]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "accSmooth[1]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "accSmooth[2]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "debug[0]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "debug[1]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "debug[2]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:     "debug[3]",
				Encoding: EncodingSignedVB,
			}, FieldDefinition{
				Name:      "motor[0]",
				Encoding:  EncodingUnsignedVB,
				Predictor: PredictorMinMotor,
			}, FieldDefinition{
				Name:      "motor[1]",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorMotor0,
			}, FieldDefinition{
				Name:      "motor[2]",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorMotor0,
			}, FieldDefinition{
				Name:      "motor[3]",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorMotor0,
			},
		},
		FieldsP: []FieldDefinition{
			FieldDefinition{
				Name:      "loopIteration",
				Encoding:  EncodingNull,
				Predictor: PredictorInc,
			}, FieldDefinition{
				Name:      "time",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorStraightLine,
			}, FieldDefinition{
				Name:      "axisP[0]",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "axisP[1]",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "axisP[2]",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "axisI[0]",
				Encoding:  EncodingTag2_3S32,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "axisI[1]",
				Encoding:  EncodingTag2_3S32,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "axisI[2]",
				Encoding:  EncodingTag2_3S32,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "axisD[0]",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "axisD[1]",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "axisF[0]",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "axisF[1]",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "axisF[2]",
				Encoding:  EncodingSignedVB,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "rcCommand[0]",
				Encoding:  EncodingTag8_4S16,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "rcCommand[1]",
				Encoding:  EncodingTag8_4S16,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "rcCommand[2]",
				Encoding:  EncodingTag8_4S16,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "rcCommand[3]",
				Encoding:  EncodingTag8_4S16,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "setpoint[0]",
				Encoding:  EncodingTag8_4S16,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "setpoint[1]",
				Encoding:  EncodingTag8_4S16,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "setpoint[2]",
				Encoding:  EncodingTag8_4S16,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "setpoint[3]",
				Encoding:  EncodingTag8_4S16,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "vbatLatest",
				Encoding:  EncodingTag8_8SVB,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "amperageLatest",
				Encoding:  EncodingTag8_8SVB,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "rssi",
				Encoding:  EncodingTag8_8SVB,
				Predictor: PredictorPrevious,
			}, FieldDefinition{
				Name:      "gyroADC[0]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "gyroADC[1]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "gyroADC[2]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "accSmooth[0]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "accSmooth[1]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "accSmooth[2]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "debug[0]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "debug[1]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "debug[2]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "debug[3]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "motor[0]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "motor[1]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "motor[2]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			}, FieldDefinition{
				Name:      "motor[3]",
				Encoding:  EncodingSignedVB,
				Predictor: PredicatorAverage2,
			},
		},
	}

	for i, field := range frameDef.FieldsP {
		if field.Encoding == EncodingTag8_8SVB {
			groupCount := 0
			for j := i + 1; j < i+8 && j < len(frameDef.FieldsP); j++ {
				groupCount = j - i
				if frameDef.FieldsP[j].Encoding != EncodingTag8_8SVB {
					break
				}
			}
			for j := i; j < i+8 && j < len(frameDef.FieldsP); j++ {
				frameDef.FieldsP[j].GroupCount = groupCount
			}

		}
	}

	frameDef.FieldIRL = map[FieldName]int{}
	for i, field := range frameDef.FieldsI {
		frameDef.FieldIRL[field.Name] = i
	}
	return frameDef
}

package blackbox

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
)

func TestReadFrameI(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFrameI))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	frame, err := frameReader.ReadNextFrame()
	assert.Nil(t, err)
	assert.Equal(t, &Frame{Type: "I", Start: 0, End: 51, Values: decodedPredictedFrameI}, frame)
}

func TestReadFrameS(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFrameS))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	frame, err := frameReader.ReadNextFrame()
	assert.Nil(t, err)
	assert.Equal(t, &Frame{Type: "S", Start: 0, End: 1, Values: []int32{}}, frame)
}

func TestReadFrameEventLoggingResume(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFramesE[0]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	frame, err := frameReader.ReadNextFrame()
	assert.Nil(t, err)
	assert.Equal(t, &Frame{Type: "E", Start: 0, End: 9, Values: nil}, frame)
	assert.Equal(t, int32(55158008), frameReader.LoggingResumeCurrentTime)
	assert.Equal(t, int32(52992), frameReader.LoggingResumeLogIteration)
}

func TestReadFrameEventSyncBeep(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFramesE[1]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	frame, err := frameReader.ReadNextFrame()
	assert.Nil(t, err)
	assert.Equal(t, &Frame{Type: "E", Start: 0, End: 6, Values: nil}, frame)
}

func TestReadFrameEventLogEnd(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFramesE[2]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	frame, err := frameReader.ReadNextFrame()
	assert.Nil(t, err)
	assert.Equal(t, &Frame{Type: "E", Start: 0, End: 12, Values: nil}, frame)
	assert.Equal(t, true, frameReader.Finished)
}

func TestReadFrameEventLogEndCorrupt(t *testing.T) {
	r := bytes.NewReader([]byte{69, 255, 'E', 'n', 'd', ' ', 'o', 'f', ' ', 'l', 'o', 'g', 'c', 'd', 'e'})
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	_, err := frameReader.ReadNextFrame()
	assert.EqualError(t, err, "There are additional data after the end of the file")
}

func TestReadStream(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFrameI, encodedFramesP[0], encodedFramesP[1], encodedFramesP[2], encodedFramesP[3], encodedFramesP[4], encodedFramesP[5]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	decodedFrames := []*Frame{
		&Frame{Type: "I", Start: 0, End: 51, Values: decodedPredictedFrameI},
		&Frame{Type: "P", Start: 51, End: 80, Values: decodedPredictedFramesP[0]},
		&Frame{Type: "P", Start: 80, End: 108, Values: decodedPredictedFramesP[1]},
		&Frame{Type: "P", Start: 108, End: 136, Values: decodedPredictedFramesP[2]},
		&Frame{Type: "P", Start: 136, End: 164, Values: decodedPredictedFramesP[3]},
		&Frame{Type: "P", Start: 164, End: 194, Values: decodedPredictedFramesP[4]},
		&Frame{Type: "P", Start: 194, End: 224, Values: decodedPredictedFramesP[5]},
	}

	for idx, decodedFrame := range decodedFrames {
		t.Run(fmt.Sprintf("for frame P%v", idx+1), func(t *testing.T) {
			frame, err := frameReader.ReadNextFrame()
			assert.Nil(t, err)
			assert.Equal(t, decodedFrame, frame)
		})
	}
}

func TestReadBrokenFrame(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFrameI, encodedFramesP[0], encodedFramesP[1], []byte{'P', 2, 3, 4}, encodedFramesP[3], encodedFramesP[4], encodedFramesP[5]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	decodedFrames := []*Frame{
		&Frame{Type: "I", Start: 0, End: 51, Values: decodedPredictedFrameI},
		&Frame{Type: "P", Start: 51, End: 80, Values: decodedPredictedFramesP[0]},
		&Frame{Type: "P", Start: 80, End: 108, Values: decodedPredictedFramesP[1]},
	}

	for idx, decodedFrame := range decodedFrames {
		t.Run(fmt.Sprintf("for frame P%v", idx+1), func(t *testing.T) {
			frame, err := frameReader.ReadNextFrame()
			assert.Nil(t, err)
			assert.Equal(t, decodedFrame, frame)
		})
	}

	_, err := frameReader.ReadNextFrame()
	assert.Nil(t, err)

	_, err = frameReader.ReadNextFrame()
	assert.EqualError(t, err, "Frame type '\v' (b'1011') is not supported")
}

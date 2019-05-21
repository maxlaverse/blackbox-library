package blackbox

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
	"github.com/stretchr/testify/assert"
)

func TestReadFrameI(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFrameI))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	frame, err := frameReader.ReadNextFrame()
	assert.Nil(t, err)
	assert.Equal(t, NewMainFrame(LogFrameIntra, decodedPredictedFrameI, 0, 51), frame)
}

func TestReadFrameS(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFrameS))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	frame, err := frameReader.ReadNextFrame()
	assert.Nil(t, err)
	assert.Equal(t, NewSlowFrame([]int32{}, 0, 1), frame)
}

func TestReadFrameEventLoggingResume(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFramesE[0]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	frame, err := frameReader.ReadNextFrame()
	assert.Nil(t, err)
	assert.Equal(t, NewEventFrame(LogEventSyncBeep, nil, 0, 9), frame)
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
	assert.Equal(t, NewEventFrame(LogEventSyncBeep, nil, 0, 6), frame)
}

func TestReadFrameEventLogEnd(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFramesE[2]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(&dec, frameDef, nil)

	frame, err := frameReader.ReadNextFrame()
	assert.Nil(t, err)
	assert.Equal(t, NewEventFrame(LogEventSyncBeep, nil, 0, 12), frame)
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

	decodedFrames := []Frame{
		NewMainFrame(LogFrameIntra, decodedPredictedFrameI, 0, 51),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[0], 51, 80),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[1], 80, 108),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[2], 108, 136),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[3], 136, 164),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[4], 164, 194),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[5], 194, 224),
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

	decodedFrames := []Frame{
		NewMainFrame(LogFrameIntra, decodedPredictedFrameI, 0, 51),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[0], 51, 80),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[1], 80, 108),
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

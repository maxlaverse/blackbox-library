package blackbox

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
	"github.com/stretchr/testify/assert"
)

func TestReadFrameI(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFrameI))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(dec, frameDef, nil)

	frame := frameReader.ReadNextFrame()
	assert.NoError(t, frame.Error())
	assert.Equal(t, valid(NewMainFrame(LogFrameIntra, decodedPredictedFrameI, 0, 51, nil)), frame)
}

func TestReadFrameS(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFrameS))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(dec, frameDef, nil)

	frame := frameReader.ReadNextFrame()
	assert.NoError(t, frame.Error())
	assert.Equal(t, valid(NewSlowFrame([]int32{}, 0, 1, nil)), frame)
}

func TestReadFrameEventLoggingResume(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFramesE[0]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(dec, frameDef, nil)

	frame := frameReader.ReadNextFrame()
	assert.NoError(t, frame.Error())

	expectedTime := int32(55158008)
	expectedIteration := int32(52992)
	expectedFrameValues := eventValues{
		"currentTime": expectedTime,
		"iteration":   expectedIteration,
		"name":        "Logging resume",
	}
	expectedFrame := NewEventFrame(LogEventLoggingResume, expectedFrameValues, 0, 9, nil)

	assert.Equal(t, valid(expectedFrame), frame)
	assert.Equal(t, expectedTime, frameReader.lastMainFrameTime)
	assert.Equal(t, expectedIteration, frameReader.lastMainFrameIteration)
}

func TestReadFrameEventSyncBeep(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFramesE[1]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(dec, frameDef, nil)

	frame := frameReader.ReadNextFrame()
	assert.NoError(t, frame.Error())

	expectedFrameValues := eventValues{
		"beepTime": uint32(41780625),
		"name":     "Sync beep",
	}
	expectedFrame := NewEventFrame(LogEventSyncBeep, expectedFrameValues, 0, 6, nil)
	assert.Equal(t, valid(expectedFrame), frame)
}

func TestReadFrameEventLogEnd(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFramesE[2]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(dec, frameDef, nil)

	frame := frameReader.ReadNextFrame()
	assert.NoError(t, frame.Error())

	expectedFrameValues := eventValues{
		"data": []byte{69, 110, 100, 32, 111, 102, 32, 108, 111, 103, 0, 0},
		"name": "Log clean end",
	}
	expectedFrame := NewEventFrame(LogEventLogEnd, expectedFrameValues, 0, 12, nil)
	assert.Equal(t, valid(expectedFrame), frame)

	frame = frameReader.ReadNextFrame()
	assert.Equal(t, io.EOF, frame.Error())
}

func TestReadFrameEventLogEndCorrupt(t *testing.T) {
	r := bytes.NewReader([]byte{69, 255, 'E', 'n', 'd', ' ', 'o', 'f', ' ', 'l', 'o', 'g', 'c', 'd', 'e'})
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(dec, frameDef, nil)

	frame := frameReader.ReadNextFrame()
	assert.EqualError(t, frame.Error(), "There are additional data after the end of the file")
}

func TestReadStream(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFrameI, encodedFramesP[0], encodedFramesP[1], encodedFramesP[2], encodedFramesP[3], encodedFramesP[4], encodedFramesP[5]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(dec, frameDef, nil)

	decodedFrames := []Frame{
		NewMainFrame(LogFrameIntra, decodedPredictedFrameI, 0, 51, nil),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[0], 51, 80, nil),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[1], 80, 108, nil),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[2], 108, 136, nil),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[3], 136, 164, nil),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[4], 164, 194, nil),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[5], 194, 224, nil),
	}

	for idx, decodedFrame := range decodedFrames {
		t.Run(fmt.Sprintf("for frame P%v", idx+1), func(t *testing.T) {
			frame := frameReader.ReadNextFrame()
			assert.NoError(t, frame.Error())
			assert.Equal(t, valid(decodedFrame), frame)
		})
	}

	frame := frameReader.ReadNextFrame()
	assert.Error(t, io.EOF, frame.Error())
}

func TestReadBrokenFrame(t *testing.T) {
	r := bytes.NewReader(buildStream(encodedFrameI, encodedFramesP[0], encodedFramesP[1], []byte{'P', 2, 3, 4}, encodedFramesP[3], encodedFramesP[4], encodedFramesP[5]))
	dec := stream.NewDecoder(r)

	frameDef := dummyFrameDefinition()

	frameReader := NewFrameReader(dec, frameDef, nil)

	decodedFrames := []Frame{
		NewMainFrame(LogFrameIntra, decodedPredictedFrameI, 0, 51, nil),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[0], 51, 80, nil),
		NewMainFrame(LogFrameInter, decodedPredictedFramesP[1], 80, 108, nil),
	}

	for idx, decodedFrame := range decodedFrames {
		t.Run(fmt.Sprintf("for frame P%v", idx+1), func(t *testing.T) {
			frame := frameReader.ReadNextFrame()
			assert.NoError(t, frame.Error())
			assert.Equal(t, valid(decodedFrame), frame)
		})
	}

	frame := frameReader.ReadNextFrame()
	assert.NoError(t, frame.Error())

	frame = frameReader.ReadNextFrame()
	assert.Equal(t, 137, frame.Start())
	assert.Equal(t, 3, frame.Size())
	assert.EqualError(t, frame.Error(), "Frame type '\v' (b'1011') is not supported")
	assert.IsType(t, &ErrorFrame{}, frame)
}

func valid(frame Frame) Frame {
	frame.setValidity(true)
	return frame
}

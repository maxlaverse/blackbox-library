package blackbox

import (
	"bytes"

	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
	"github.com/pkg/errors"
)

const (
	// maximumIterationJumpBetweenFrames is the maximum number of frames that can be missing without
	// declaring the stream as invalid
	maximumIterationJumpBetweenFrames = 5000

	// maximumTimeJumpBetweenFrames is the maximum delay in microseconds between two consecutive main
	// frames before the stream is declared invalid
	maximumTimeJumpBetweenFrames = 10 * 1000000

	// maxFrameLength is the maximum size of a frame. Bigger means something is off
	maxFrameLength = 256
)

// FieldName is the type for main frame fields
type FieldName string

// List of all the fields used by the library
const (
	FieldIteration        FieldName = "loopIteration"
	FieldTime             FieldName = "time"
	FieldVbatLatest       FieldName = "vbatLatest"
	FieldAmperageLatest   FieldName = "amperageLatest"
	FieldEnergyCumulative FieldName = "energyCumulative"
	FieldFlightModeFlags  FieldName = "flightModeFlags"
	FieldStateFlags       FieldName = "stateFlags"
	FieldFailsafePhase    FieldName = "failsafePhase"
	FieldMotor0           FieldName = "motor[0]"
)

// FrameReader reads and decodes data frame
type FrameReader struct {
	timeRolloverAccumulator int64
	lastMainFrameIteration  int64
	lastMainFrameTime       int64
	mainStreamIsValid       bool
	previousFrame1          *MainFrame
	previousFrame2          *MainFrame
	dec                     *stream.Decoder
	frameDef                LogDefinition
	opts                    FrameReaderOptions
}

// FrameReaderOptions holds the options to create a new FrameReader
type FrameReaderOptions struct {
	Raw bool
}

// NewFrameReader returns a new FrameReader
func NewFrameReader(dec *stream.Decoder, frameDef LogDefinition, opts *FrameReaderOptions) *FrameReader {
	if opts == nil {
		opts = &FrameReaderOptions{}
	}

	return &FrameReader{
		dec:                    dec,
		frameDef:               frameDef,
		lastMainFrameIteration: -1,
		lastMainFrameTime:      -1,
		opts:                   *opts,
	}
}

// ReadNextFrame reads the next frame
func (f *FrameReader) ReadNextFrame() Frame {
	startOffset := f.dec.Offset()

	// Determine the type of the next frame
	frameType, err := f.dec.ReadByte()
	if err != nil {
		return NewErrorFrame(nil, startOffset, f.dec.Offset(), err)
	}

	// Read the frame
	frame := f.parseFrame(frameType, startOffset)

	// Check if the frame could be read and as at least a reasonable size
	if frame.Error() != nil && frame.Size() > maxFrameLength {
		frame.setError(frameErrorLengthLimit(frame.Size()))
		return frame
	}

	// Verify the frame is actually alright and update the reader's state
	frameAccepted := f.validateFrame(frame)
	frame.setValidity(frameAccepted)
	return frame
}

// parseFrame reads bytes until a frame is complete
func (f *FrameReader) parseFrame(frameType byte, startOffset int64) Frame {
	switch frameType {
	case LogFrameEvent:
		eventType, values, err := parseEventFrame(f.dec)
		return NewEventFrame(eventType, values, startOffset, f.dec.Offset(), err)

	case LogFrameSlow:
		values, err := parseStateFrame(f.frameDef, f.frameDef.FieldsS, nil, nil, f.dec, f.opts.Raw, 0)
		return NewSlowFrame(values, startOffset, f.dec.Offset(), err)

	case LogFrameIntra:
		values, err := parseStateFrame(f.frameDef, f.frameDef.FieldsI, f.previousFrame1, f.previousFrame2, f.dec, f.opts.Raw, 0)
		return NewMainFrame(frameType, values, startOffset, f.dec.Offset(), err)

	case LogFrameInter:
		values, err := parseStateFrame(f.frameDef, f.frameDef.FieldsP, f.previousFrame1, f.previousFrame2, f.dec, f.opts.Raw, f.countIntentionallySkippedFrames())
		return NewMainFrame(frameType, values, startOffset, f.dec.Offset(), err)

	default:
		values, err := f.readBytesToNextFrame()
		return NewErrorFrame(values, startOffset, f.dec.Offset(), errors.WithStack(frameErrorUnsupportedType(frameType, err)))
	}
}

// validateFrame checks if a frame is valid and update the reader's internal state
func (f *FrameReader) validateFrame(frame Frame) bool {
	switch frame.Type() {
	case LogFrameEvent:
		if frame.(*EventFrame).EventType() == LogEventLoggingResume {
			f.lastMainFrameIteration = frame.Values().(map[string]interface{})["iteration"].(int64)
			f.lastMainFrameTime = frame.Values().(map[string]interface{})["currentTime"].(int64) + f.timeRolloverAccumulator
		}
		return true

	case LogFrameSlow:
		return true

	case LogFrameIntra:
		f.flightLogApplyMainFrameTimeRollover(frame.(*MainFrame))

		if !f.opts.Raw && f.lastMainFrameIteration != -1 && !f.validateMainFrameValues(frame.(*MainFrame)) {
			f.mainStreamIsValid = false
			f.previousFrame1 = nil
			f.previousFrame2 = nil
		} else {
			f.mainStreamIsValid = true
		}

		if f.mainStreamIsValid {
			f.lastMainFrameIteration = frame.(*MainFrame).values[0]
			f.lastMainFrameTime = frame.(*MainFrame).values[1]

			// Rotate history buffers
			f.previousFrame2 = frame.(*MainFrame)
			f.previousFrame1 = frame.(*MainFrame)
		}
		return f.mainStreamIsValid

	case LogFrameInter:
		f.flightLogApplyMainFrameTimeRollover(frame.(*MainFrame))

		// Only attempt to validate the frame values if we have something to check it against
		if !f.opts.Raw && f.mainStreamIsValid && !f.validateMainFrameValues(frame.(*MainFrame)) {
			f.mainStreamIsValid = false
			f.previousFrame1 = nil
			f.previousFrame2 = nil
		}

		if f.mainStreamIsValid {
			// TODO: Remove hard-coded field indexes
			f.lastMainFrameIteration = frame.(*MainFrame).values[0]
			f.lastMainFrameTime = frame.(*MainFrame).values[1]

			// Rotate history buffers
			f.previousFrame2 = f.previousFrame1
			f.previousFrame1 = frame.(*MainFrame)
		}

		return f.mainStreamIsValid
	}
	return false
}

// readBytesToNextFrame reads bytes until the begining of a frame is found or the end of the file is reached
func (f *FrameReader) readBytesToNextFrame() ([]byte, error) {
	values := []byte{}
	for {
		b, err := f.dec.NextByte()
		if err != nil || bytes.IndexByte(LogFrameAllTypes, b) != -1 {
			return values, err
		}

		b, err = f.dec.ReadByte()
		if err != nil {
			return values, err
		}
		values = append(values, b)
	}
}

func (f *FrameReader) countIntentionallySkippedFrames() int64 {
	if f.lastMainFrameIteration == -1 {
		return 0
	}

	count := 0
	for frameIndex := f.lastMainFrameIteration + 1; !f.frameExpected(int(frameIndex)); frameIndex++ {
		count++
	}

	return int64(count)
}

func (f *FrameReader) frameExpected(frameIndex int) bool {
	return (frameIndex%f.frameDef.Sysconfig.FrameIntervalI+f.frameDef.Sysconfig.FrameIntervalPNum-1)%f.frameDef.Sysconfig.FrameIntervalPDenom < f.frameDef.Sysconfig.FrameIntervalPNum
}

func (f *FrameReader) flightLogApplyMainFrameTimeRollover(frame *MainFrame) {
	frame.values[1] = f.flightLogDetectAndApplyTimestampRollover(frame.values[1])
}

func (f *FrameReader) flightLogDetectAndApplyTimestampRollover(timestamp int64) int64 {
	if f.lastMainFrameTime != -1 && uint32(timestamp) < uint32(f.lastMainFrameTime) && uint32(uint32(timestamp)-uint32(f.lastMainFrameTime)) < maximumTimeJumpBetweenFrames {
		f.timeRolloverAccumulator += 4294967296
	}
	return timestamp + f.timeRolloverAccumulator
}

// validateMainFrameValues verifies if the frame's loopIteration and time makes sense
func (f *FrameReader) validateMainFrameValues(frame *MainFrame) bool {
	//TODO: Remove hard-coded indexes
	return frame.values[0] >= f.lastMainFrameIteration &&
		frame.values[0] < f.lastMainFrameIteration+maximumIterationJumpBetweenFrames &&
		frame.values[1] >= f.lastMainFrameTime &&
		frame.values[1] < f.lastMainFrameTime+maximumTimeJumpBetweenFrames
}

func frameErrorLengthLimit(size int) error {
	return errors.Errorf("frame size %d is bigger than the maximum allowed value %d", size, maxFrameLength)
}

func frameErrorUnsupportedType(frameType LogFrameType, err error) error {
	if err != nil {
		return errors.Errorf("Frame type '%s' (b'%b') is not supported: %v", string(frameType), frameType, err)
	}
	return errors.Errorf("Frame type '%s' (b'%b') is not supported", string(frameType), frameType)
}

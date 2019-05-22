package blackbox

import (
	"strconv"

	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
	"github.com/pkg/errors"
)

const (
	maximumIterationJumpBetweenFrames = 5000
	maximumTimeJumpBetweenFrames      = 10 * 1000000
	logEventMaxFrameLength            = 256
)

type FieldName string

// List of all the fields used by the library
const (
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
	LoggingResumeLogIteration int32
	LastEventType             *LogEventType
	LastSkippedFrames         int32
	LoggingResumeCurrentTime  int32
	timeRolloverAccumulator   int32
	Finished                  bool
	LastMainFrameIteration    int32
	LastMainFrameTime         int32
	MainStreamIsValid         bool
	Previous                  *MainFrame
	PreviousPrevious          *MainFrame
	frameIntervalI            int
	frameIntervalPNum         int
	frameIntervalPDenom       int
	Stats                     *StatsType
	dec                       *stream.Decoder
	frameDef                  LogDefinition
	raw                       bool
}

// FrameReaderOptions holds the options to create a new FrameReader
type FrameReaderOptions struct {
	Raw bool
}

// NewFrameReader returns a new FrameReader
func NewFrameReader(dec *stream.Decoder, frameDef LogDefinition, opts *FrameReaderOptions) (*FrameReader, error) {
	val, err := frameDef.GetHeaderValue(HeaderIInterval)
	if err != nil {
		return nil, err
	}
	frameIntervalI, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return nil, err
	}
	if frameIntervalI < 1 {
		frameIntervalI = 1
	}

	//TODO: Set frameIntervalPNum and frameIntervalPDenom based on "P Interval" header

	frameReader := FrameReader{
		dec:                 dec,
		frameDef:            frameDef,
		frameIntervalI:      int(frameIntervalI),
		frameIntervalPNum:   1,
		frameIntervalPDenom: 1,
		Stats: &StatsType{
			Frame: map[LogFrameType]StatsFrameType{},
		},
		LastMainFrameIteration: -1,
		LastMainFrameTime:      -1,
		LastSkippedFrames:      0,
	}
	if opts != nil {
		frameReader.raw = opts.Raw
	}
	return &frameReader, nil
}

// ReadNextFrame reads the next frame
func (f *FrameReader) ReadNextFrame() Frame {
	// create some base frame to make sure we always have something to return
	// even if we could not read any frame data
	var frame Frame = &baseFrame{}

	start := f.dec.BytesRead()
	frameType, err := f.dec.ReadByte()
	if err != nil {
		frame.addError(err)
		f.countFrameAsCorrupted(frame)
		return frame
	}

	// parse next frame data based on its type
	switch frameType {
	case LogFrameEvent:
		eventType, eventValues, err := f.parseEventFrame(f.dec)
		frame = NewEventFrame(eventType, eventValues, start, f.dec.BytesRead())
		frame.addError(err)

	case LogFrameIntra:
		values, err := ParseFrame(f.frameDef, f.frameDef.FieldsI, f.Previous, f.PreviousPrevious, f.dec, f.raw, f.LastSkippedFrames)
		frame = NewMainFrame(frameType, values, start, f.dec.BytesRead())
		frame.addError(err)

	case LogFrameInter:
		f.LastSkippedFrames = f.countIntentionallySkippedFrames()
		values, err := ParseFrame(f.frameDef, f.frameDef.FieldsP, f.Previous, f.PreviousPrevious, f.dec, f.raw, f.LastSkippedFrames)
		frame = NewMainFrame(frameType, values, start, f.dec.BytesRead())
		frame.addError(err)

	case LogFrameSlow:
		values, err := ParseFrame(f.frameDef, f.frameDef.FieldsS, nil, nil, f.dec, f.raw, 0)
		frame = NewSlowFrame(values, start, f.dec.BytesRead())
		frame.addError(err)

	default:
		frame.addError(frameErrorUnsupportedType(frameType))
		f.countFrameAsCorrupted(frame)
		//TODO: If a P frame is corrupt here, we should optionally drop the previous one (As the original implementation does)
		f.flightLogReaderInvalidateStream()
		return frame
	}

	f.PreComplete(frame)
	// if frame has any reading/parsing errors - it should be marked as corrupted
	if !frame.IsValid() {
		f.countFrameAsCorrupted(frame)
	}

	return frame
}

func (f *FrameReader) parseEventFrame(dec *stream.Decoder) (LogEventType, eventValues, error) {
	values := make(eventValues)
	eventType, err := dec.ReadByte()
	if err != nil {
		return 0, nil, err
	}
	f.LastEventType = &eventType

	switch eventType {
	case LogEventSyncBeep:
		beepTime, err := dec.ReadUnsignedVB()
		if err != nil {
			return 0, nil, err
		}
		values["beepTime"] = beepTime

	case LogEventInflightAdjustment:
		return 0, nil, errors.New("Not implemented: logEventInflightAdjustment")

	case LogEventLoggingResume:
		val, err := dec.ReadUnsignedVB()
		if err != nil {
			return 0, nil, err
		}
		f.LoggingResumeLogIteration = int32(val)

		val, err = dec.ReadUnsignedVB()
		if err != nil {
			return 0, nil, err
		}
		f.LoggingResumeCurrentTime = int32(val) + f.timeRolloverAccumulator
		values["iteration"] = f.LoggingResumeLogIteration
		values["currentTime"] = f.LoggingResumeCurrentTime

	case LogEventFlightMode:
		flags, err := dec.ReadUnsignedVB()
		if err != nil {
			return 0, nil, err
		}
		lastFlags, err := dec.ReadUnsignedVB()
		if err != nil {
			return 0, nil, err
		}
		values["flags"] = flags
		values["lastFlags"] = lastFlags

	case LogEventLogEnd:
		val, err := dec.ReadBytes(12)
		if err != nil {
			return 0, nil, err
		}

		reachedEndOfFile, err := dec.EOF()
		if err != nil {
			return 0, nil, err
		}
		if !reachedEndOfFile {
			return 0, nil, errors.New("There are additional data after the end of the file")
		}
		f.Finished = true
		values["data"] = val

	default:
		return 0, nil, errors.Errorf("Event type is unknown - ignored: %v\n", eventType)
	}

	return eventType, values, nil
}

// PreComplete updates the reading state machine based on a frame
func (f *FrameReader) PreComplete(frame Frame) (frameAccepted bool) {
	f.Stats.TotalFrames++

	if !frame.IsRead() {
		return false
	}

	if frame.Size() > logEventMaxFrameLength {
		frame.addError(frameErrorLengthLimit(frame.Size()))
		return false
	}

	frameAccepted = f.Complete(frame)

	if frameAccepted {
		d := f.Stats.Frame[frame.Type()]
		d.Bytes += int64(frame.Size())
		d.ValidCount++
		if d.SizeCount == nil {
			d.SizeCount = map[int]int{}
		}
		d.SizeCount[frame.Size()]++
		f.Stats.Frame[frame.Type()] = d
	}

	d := f.Stats.Frame[frame.Type()]
	d.DesyncCount++
	f.Stats.Frame[frame.Type()] = d

	return frameAccepted
}

// Complete acknowledge and saves a frame
func (f *FrameReader) Complete(frame Frame) (frameCompleted bool) {
	switch frame.Type() {
	case LogFrameIntra:
		return f.completeFrameIntra(frame.(*MainFrame))
	case LogFrameInter:
		return f.completeFrameInter(frame.(*MainFrame))
	case LogFrameSlow:
		return f.completeSlowFrame(frame.(*SlowFrame))
	case LogFrameEvent:
		return f.completeEventFrame()
	default:
		frame.addError(frameErrorUnsupportedType(frame.Type()))
		return false
	}
}

func (f *FrameReader) countFrameAsCorrupted(frame Frame) {
	f.MainStreamIsValid = false
	d := f.Stats.Frame[frame.Type()]
	d.CorruptCount++
	f.Stats.Frame[frame.Type()] = d
	f.Stats.TotalCorruptedFrames++
}

func (f *FrameReader) completeFrameIntra(frame *MainFrame) bool {
	// Change time
	f.flightLogApplyMainFrameTimeRollover(frame)

	// Only attempt to validate the frame values if we have something to check it against
	if f.LastMainFrameIteration != -1 && !f.verifyFrameMainValues(frame) {
		f.flightLogReaderInvalidateStream()
	} else {
		f.MainStreamIsValid = true
	}

	if f.MainStreamIsValid {
		f.Stats.IntentionallyAbsentIterations += int(f.countIntentionallySkippedFramesTo(frame.values[0]))

		f.LastMainFrameIteration = frame.values[0]
		f.LastMainFrameTime = frame.values[1]
	}

	if f.MainStreamIsValid {
		//TODO: Append the frame to an array, and push bash through a channel
		f.Previous = frame
		f.PreviousPrevious = frame
	}
	return f.MainStreamIsValid
}

func (f *FrameReader) completeFrameInter(frame *MainFrame) bool {
	f.flightLogApplyMainFrameTimeRollover(frame)

	if f.LastMainFrameIteration != -1 && !f.verifyFrameMainValues(frame) {
		f.flightLogReaderInvalidateStream()
	}

	if f.MainStreamIsValid {
		f.LastMainFrameIteration = frame.values[0]
		f.LastMainFrameTime = frame.values[1]
		f.Stats.IntentionallyAbsentIterations += int(f.LastSkippedFrames)
	}

	// Receiving a P frame can't resynchronise the stream so it doesn't set mainStreamIsValid to true
	if f.MainStreamIsValid {
		f.PreviousPrevious = f.Previous
		f.Previous = frame
	}
	return f.MainStreamIsValid
}

func (f *FrameReader) completeSlowFrame(frame *SlowFrame) bool {
	return true
}

func (f *FrameReader) completeEventFrame() bool {
	if f.LastEventType != nil {
		if *f.LastEventType == LogEventLoggingResume {
			f.LastMainFrameIteration = f.LoggingResumeLogIteration
			f.LastMainFrameTime = f.LoggingResumeCurrentTime
		}
		f.MainStreamIsValid = true
		return true
	}
	return false
}

func (f *FrameReader) flightLogReaderInvalidateStream() {
	f.MainStreamIsValid = false
	f.Previous = nil
	f.PreviousPrevious = nil
}

func (f *FrameReader) countIntentionallySkippedFramesTo(targetIteration int32) int32 {
	count := 0

	if f.LastMainFrameIteration == -1 {
		// Haven't parsed a frame yet so there's no frames to skip
		return 0
	}

	for frameIndex := f.LastMainFrameIteration + 1; frameIndex < targetIteration; frameIndex++ {
		if f.shouldHaveFrame(int(frameIndex)) {
			count++
		}
	}

	return int32(count)
}

func (f *FrameReader) flightLogApplyMainFrameTimeRollover(frame *MainFrame) {
	frame.values[1] = f.flightLogDetectAndApplyTimestampRollover(frame.values[1])
}

func (f *FrameReader) flightLogDetectAndApplyTimestampRollover(timestamp int32) int32 {
	if f.LastMainFrameTime != -1 {
		if timestamp < f.LastMainFrameTime && timestamp-f.LastMainFrameTime < maximumTimeJumpBetweenFrames {
			f.timeRolloverAccumulator = 2147483647
		}
	}

	return timestamp + f.timeRolloverAccumulator
}

// verifyFrameMainValues verifies if the frame's loopIteration and time makes sense
func (f *FrameReader) verifyFrameMainValues(frame *MainFrame) bool {
	return frame.values[0] >= f.LastMainFrameIteration &&
		frame.values[0] < f.LastMainFrameIteration+maximumIterationJumpBetweenFrames &&
		frame.values[1] >= f.LastMainFrameTime &&
		frame.values[1] < f.LastMainFrameTime+maximumTimeJumpBetweenFrames
}

func (f *FrameReader) countIntentionallySkippedFrames() int32 {
	count := 0
	if f.LastMainFrameIteration == -1 {
		// Haven't parsed a frame yet so there's no frames to skip
		return 0
	}

	for frameIndex := f.LastMainFrameIteration + 1; !f.shouldHaveFrame(int(frameIndex)); frameIndex++ {
		count++
	}

	return int32(count)
}

func (f *FrameReader) shouldHaveFrame(frameIndex int) bool {
	return (frameIndex%f.frameIntervalI+f.frameIntervalPNum-1)%f.frameIntervalPDenom < f.frameIntervalPNum
}

func frameErrorLengthLimit(size int) error {
	return errors.Errorf(
		"FrameReader: previous frame was corrupt, main stream is not valid anymore, frame size %d is bigger than expected %d",
		size,
		logEventMaxFrameLength,
	)
}

func frameErrorUnsupportedType(frameType LogFrameType) error {
	return errors.Errorf("Frame type '%s' (b'%b') is not supported", string(frameType), frameType)
}

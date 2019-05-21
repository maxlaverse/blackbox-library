package blackbox

import (
	"fmt"
	"os"
	"strconv"

	"text/tabwriter"

	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
	"github.com/pkg/errors"
)

const (
	maximumIterationJumpBetweenFrames = 5000
	maximumTimeJumpBetweenFrames      = 10 * 1000000
	logEventMaxFrameLength            = 256
)

// FrameReader reads and decodes data frame
type FrameReader struct {
	LoggingResumeLogIteration int32
	LastEventType             LogEventType
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
	Stats                     StatsType
	dec                       *stream.Decoder
	frameDef                  LogDefinition
	raw                       bool
}

// FrameReaderOptions holds the options to create a new FrameReader
type FrameReaderOptions struct {
	Raw bool
}

// NewFrameReader returns a new FrameReader
func NewFrameReader(dec *stream.Decoder, frameDef LogDefinition, opts *FrameReaderOptions) FrameReader {
	val, err := frameDef.GetHeaderValue("I interval")
	if err != nil {
		panic(err)
	}
	frameIntervalI, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		panic(err)
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
		Stats: StatsType{
			Frame: map[LogFrameType]StatsFrameType{},
		},
		LastMainFrameIteration: -1,
		LastMainFrameTime:      -1,
		LastSkippedFrames:      0,
	}
	if opts != nil {
		frameReader.raw = opts.Raw
	}
	return frameReader
}

// ReadNextFrame reads the next frame
func (f *FrameReader) ReadNextFrame() (Frame, error) {
	start := f.dec.BytesRead()

	command, err := f.dec.ReadByte()
	if err != nil {
		return nil, err
	}
	frameType := LogFrameType(command)

	//Next byte should be readable if stream is valid
	switch frameType {
	case LogFrameEvent:
		err := f.parseEventFrame(f.dec)
		if err != nil {
			return nil, err
		}
		end := f.dec.BytesRead()
		// @TODO: parse event properly
		frame := NewEventFrame(LogEventSyncBeep, nil, start, end)
		f.PreComplete(frame)
		return frame, nil

	case LogFrameIntra:
		values, err := ParseFrame(f.frameDef, f.frameDef.FieldsI, f.Previous, f.PreviousPrevious, f.dec, f.raw, f.LastSkippedFrames)
		if err != nil {
			return nil, err
		}
		end := f.dec.BytesRead()
		frame := NewMainFrame(frameType, values, start, end)
		f.PreComplete(frame)
		return frame, nil

	case LogFrameInter:
		f.LastSkippedFrames = f.countIntentionallySkippedFrames()
		values, err := ParseFrame(f.frameDef, f.frameDef.FieldsP, f.Previous, f.PreviousPrevious, f.dec, f.raw, f.LastSkippedFrames)
		if err != nil {
			return nil, err
		}
		end := f.dec.BytesRead()
		frame := NewMainFrame(frameType, values, start, end)
		f.PreComplete(frame)
		return frame, nil

	case LogFrameSlow:
		values, err := ParseFrame(f.frameDef, f.frameDef.FieldsS, nil, nil, f.dec, f.raw, 0)
		if err != nil {
			return nil, err
		}
		end := f.dec.BytesRead()
		frame := NewSlowFrame(values, start, end)
		f.PreComplete(frame)
		return frame, nil

	default:
		//TODO: If a P frame is corrupt here, we should optionally drop the previous one (As the original implementation does)
		f.flightLogReaderInvalidateStream()
		f.Stats.TotalCorruptedFrames++
		return nil, errors.Errorf("Frame type '%s' (b'%b') is not supported", string(command), command)
	}
}

func (f *FrameReader) parseEventFrame(dec *stream.Decoder) error {
	event, err := dec.ReadByte()
	if err != nil {
		return err
	}
	f.LastEventType = LogEventType(event)

	switch f.LastEventType {
	case LogEventSyncBeep:
		beepTime, err := dec.ReadUnsignedVB()
		if err != nil {
			return err
		}
		fmt.Printf("Frame #X (E) ")
		fmt.Printf("%s=%d ", "beepTime", beepTime)
		fmt.Printf("\n")
	case LogEventInflightAdjustment:
		fmt.Printf("Event type is logEventInflightAdjustment\n")
		panic("Not implemented: logEventInflightAdjustment")
	case LogEventLoggingResume:
		val, err := dec.ReadUnsignedVB()
		if err != nil {
			return err
		}
		f.LoggingResumeLogIteration = int32(val)

		val, err = dec.ReadUnsignedVB()
		if err != nil {
			return err
		}
		f.LoggingResumeCurrentTime = int32(val) + f.timeRolloverAccumulator

		fmt.Printf("Frame #X (%s) ", "E")
		fmt.Printf("%s=%d ", "iteration", f.LoggingResumeLogIteration)
		fmt.Printf("%s=%d ", "currentTime", f.LoggingResumeCurrentTime)
		fmt.Printf("\n")
	case LogEventFlightMode:
		fmt.Printf("Frame #X (%s) ", "E")
		flags, err := dec.ReadUnsignedVB()
		if err != nil {
			return err
		}
		fmt.Printf("%s=%d ", "flags", flags)

		lastFlags, err := dec.ReadUnsignedVB()
		if err != nil {
			return err
		}
		fmt.Printf("%s=%d ", "lastFlags", lastFlags)
		fmt.Printf("\n")
	case LogEventLogEnd:

		val, err := dec.ReadBytes(12)
		if err != nil {
			return err
		}
		fmt.Printf("Frame #X (%s) ", "E")
		fmt.Printf("data=%s ", val)
		fmt.Printf("\n")
		reachedEndOfFile, err := dec.EOF()
		if err != nil {
			return err
		}
		if !reachedEndOfFile {
			return errors.New("There are additional data after the end of the file")
		}
		f.Finished = true
	default:
		fmt.Printf("Event type is unknown - ignored: %v\n", event)
	}
	return nil
}

// PreComplete updates the reading state machine based on a frame
func (f *FrameReader) PreComplete(frame Frame) bool {
	f.Stats.TotalFrames++
	if frame.Size() <= logEventMaxFrameLength {
		frameAccepted := f.Complete(frame)
		if frameAccepted {
			d := f.Stats.Frame[frame.Type()]
			d.Bytes += int64(frame.Size())
			d.ValidCount++
			if d.SizeCount == nil {
				d.SizeCount = map[int]int{}
			}
			d.SizeCount[frame.Size()]++
			f.Stats.Frame[frame.Type()] = d
			return true
		}

		d := f.Stats.Frame[frame.Type()]
		d.DesyncCount++
		f.Stats.Frame[frame.Type()] = d
		return false

	}

	fmt.Printf("The previous frame was corrupt - Main stream is not valid anymore\n")
	f.MainStreamIsValid = false
	d := f.Stats.Frame[frame.Type()]
	d.CorruptCount++
	f.Stats.Frame[frame.Type()] = d
	f.Stats.TotalCorruptedFrames++
	panic(frame.Size())
}

// Complete aknowledge and saves a frame
func (f *FrameReader) Complete(frame Frame) bool {
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
		panic(fmt.Sprintf("Unable to complete '%s'", frame.Type()))
	}
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
	if f.LastEventType != -1 {
		if f.LastEventType == LogEventLoggingResume {
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

// PrintStatistics prints statistics on the log
func (f *FrameReader) PrintStatistics() {
	fmt.Println("Statistics")
	w1 := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w1, "TotalFrames\t %d\n", f.Stats.TotalFrames)
	fmt.Fprintf(w1, "TotalCorruptedFrames\t %d\n", f.Stats.TotalCorruptedFrames)
	w1.Flush()

	fmt.Println("Statistics")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	for t, f := range f.Stats.Frame {
		fmt.Fprintf(w, "%s frames\t %d valid\t %d desync\t %d bytes\t %d corrupt\t %d sizes\t\n", t, f.ValidCount, f.DesyncCount, f.Bytes, f.CorruptCount, len(f.SizeCount))
	}
	w.Flush()
}

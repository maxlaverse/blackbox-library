package blackbox

import "github.com/maxlaverse/blackbox-library/src/blackbox/stream"

// See https://cleanflight.readthedocs.io/en/stable/development/Blackbox%20Internals/

type Frame interface {
	Type() LogFrameType
	Values() interface{}
	Sized
	ErroredFrame
}

type Sized interface {
	Size() int
}

type ErroredFrame interface {
	// IsRead tells if there weren't errors while reading the frame from data source
	IsRead() bool
	// IsValid tells if there weren't errors while parsing the read data for this frame
	IsValid() bool
	// Errors returns all errors which happened while reading and parsing data for this frame
	Errors() []error
	// addError adds an error to the frame. If error is nil then does nothing
	addError(err error)
}

// -------------------------------------------------------------------------- //

type baseFrame struct {
	frameType  LogFrameType
	values     interface{}
	errors     []error
	isMissRead bool
	start      int64
	end        int64
}

func (f baseFrame) Type() LogFrameType {
	return f.frameType
}

func (f baseFrame) Values() interface{} {
	return f.values
}

// Size returns the size in bytes of a Frame
func (f baseFrame) Size() int {
	return int(f.end - f.start)
}

func (f baseFrame) Errors() []error {
	return f.errors
}

func (f baseFrame) IsRead() bool {
	return !f.isMissRead
}

func (f baseFrame) IsValid() bool {
	return len(f.errors) == 0
}

func (f *baseFrame) addError(err error) {
	if err == nil {
		return
	}
	if _, ok := err.(stream.ReadError); ok {
		f.isMissRead = true
	}
	f.errors = append(f.errors, err)
}

// -------------------------------------------------------------------------- //

// Frame represents a frame
type MainFrame struct {
	baseFrame
	values []int32
}

// NewFrame returns a new frame
func NewMainFrame(frameType LogFrameType, values []int32, start, end int64) *MainFrame {
	return &MainFrame{
		values: values,
		baseFrame: baseFrame{
			frameType: frameType,
			start:     start,
			end:       end,
		},
	}
}

func (f MainFrame) Values() interface{} {
	return f.values
}

// -------------------------------------------------------------------------- //

// Frame represents a frame
type SlowFrame struct {
	baseFrame
	values []int32
}

// NewFrame returns a new frame
func NewSlowFrame(values []int32, start, end int64) *SlowFrame {
	return &SlowFrame{
		values: values,
		baseFrame: baseFrame{
			frameType: LogFrameSlow,
			start:     start,
			end:       end,
		},
	}
}

func (f SlowFrame) Values() interface{} {
	return f.values
}

// -------------------------------------------------------------------------- //

type eventValues = map[string]interface{}

type EventFrame struct {
	baseFrame
	eventType LogEventType
	values    eventValues
}

func NewEventFrame(eventType LogEventType, values eventValues, start, end int64) *EventFrame {
	return &EventFrame{
		eventType: eventType,
		values:    values,
		baseFrame: baseFrame{
			frameType: LogFrameEvent,
			start:     start,
			end:       end,
		},
	}
}

func (f EventFrame) Values() interface{} {
	return f.values
}

func (f EventFrame) EventType() LogEventType {
	return f.eventType
}

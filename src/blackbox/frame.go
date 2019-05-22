package blackbox

// See https://cleanflight.readthedocs.io/en/stable/development/Blackbox%20Internals/

type Frame interface {
	Type() LogFrameType
	Values() interface{}
	Sizer
	Errorer
}

type Sizer interface {
	Size() int
}

type Errorer interface {
	// Error returns error which happened during reading/parsing the frame, or nil
	Error() error
	// setError adds the error to the frame
	setError(err error)
}

// -------------------------------------------------------------------------- //

type baseFrame struct {
	frameType LogFrameType
	values    interface{}
	error     error
	start     int64
	end       int64
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

func (f baseFrame) Error() error {
	return f.error
}

func (f *baseFrame) setError(err error) {
	f.error = err
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

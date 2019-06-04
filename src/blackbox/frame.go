package blackbox

import (
	"fmt"
	"sort"
	"strings"
)

var flightModeNames = map[int32]string{
	1:   "ANGLE_MODE",
	2:   "HORIZON_MODE",
	4:   "MAG",
	8:   "BARO",
	16:  "GPS_HOME",
	32:  "GPS_HOLD",
	64:  "HEADFREE",
	128: "AUTOTUNE",
	256: "PASSTHRU",
	512: "SONAR",
}

var flightStateNames = map[int32]string{
	1:  "GPS_FIX_HOME",
	2:  "GPS_FIX",
	4:  "CALIBRATE_MAG",
	8:  "SMALL_ANGLE",
	16: "FIXED_WING",
}

var failsafePhaseNames = []string{
	"IDLE",
	"RX_LOSS_DETECTED",
	"LANDING",
	"LANDED",
}

// See https://cleanflight.readthedocs.io/en/stable/development/Blackbox%20Internals/

type Frame interface {
	Type() LogFrameType
	Values() interface{}
	Sizer
	Errorer
	Validater
}

type Sizer interface {
	Size() int
	Start() int
}

type Errorer interface {
	// Error returns error which happened during reading/parsing the frame, or nil
	Error() error
	// setError adds the error to the frame
	setError(err error)
}

type Validater interface {
	Validity() bool
	setValidity(validity bool)
}

// -------------------------------------------------------------------------- //

type baseFrame struct {
	frameType LogFrameType
	values    interface{}
	err       error
	start     int64
	end       int64
	validity  bool
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

// Start returns the Start in bytes of a Frame
func (f baseFrame) Start() int {
	return int(f.start)
}

func (f baseFrame) Error() error {
	return f.err
}

func (f *baseFrame) setError(err error) {
	f.err = err
}

func (f baseFrame) Validity() bool {
	return f.validity
}

func (f *baseFrame) setValidity(validity bool) {
	f.validity = validity
}

// -------------------------------------------------------------------------- //

// Frame represents a frame
type MainFrame struct {
	baseFrame
	values []int32
}

// NewFrame returns a new frame
func NewMainFrame(frameType LogFrameType, values []int32, start, end int64, err error) *MainFrame {
	return &MainFrame{
		values: values,
		baseFrame: baseFrame{
			frameType: frameType,
			start:     start,
			end:       end,
			err:       err,
		},
	}
}

func (f MainFrame) Values() interface{} {
	return f.values
}

// -------------------------------------------------------------------------- //

// SlowFrame represents a slow frame
type SlowFrame struct {
	baseFrame
	values []int32
}

// NewSlowFrame returns a new frame
func NewSlowFrame(values []int32, start, end int64, err error) *SlowFrame {
	return &SlowFrame{
		values: values,
		baseFrame: baseFrame{
			frameType: LogFrameSlow,
			start:     start,
			end:       end,
			err:       err,
		},
	}
}

// ErrorFrame represents a frame that couldn't be recognized
type ErrorFrame struct {
	baseFrame
	values []byte
}

func NewErrorFrame(values []byte, start, end int64, err error) *ErrorFrame {
	return &ErrorFrame{
		values: values,
		baseFrame: baseFrame{
			frameType: 0,
			start:     start,
			end:       end,
			err:       err,
		},
	}
}

func (f SlowFrame) Values() interface{} {
	return f.values
}

func (f SlowFrame) StringValues() []string {
	values := make([]string, len(f.values))
	for k, v := range f.values {
		values[k] = slowFrameFlagToString(k, v)
	}
	return values
}

func (f SlowFrame) String() string {
	return fmt.Sprintf("S frame: %s", strings.Join(f.StringValues(), ", "))
}

func slowFrameFlagToString(fieldIndex int, value int32) string {
	switch fieldIndex {
	case 0:
		return decodeFlagsToString(flightModeNames, value)
	case 1:
		return decodeFlagsToString(flightStateNames, value)
	case 2:
		return decodeEnumToString(failsafePhaseNames, value)
	default:
		return fmt.Sprintf("%d", value)
	}
}

func decodeFlagsToString(flags map[int32]string, flagValue int32) string {
	flagsAsStrings := []string{}
	for k, v := range flags {
		if flagValue&k == k {
			flagsAsStrings = append(flagsAsStrings, v)
		}
	}
	if len(flagsAsStrings) == 0 {
		return "0"
	}
	return strings.Join(flagsAsStrings, "|")
}

func decodeEnumToString(enum []string, value int32) string {
	if int(value) >= len(enum) {
		return fmt.Sprintf("%d", value)
	}
	return enum[value]
}

// -------------------------------------------------------------------------- //

type eventValues = map[string]interface{}

type EventFrame struct {
	baseFrame
	eventType LogEventType
	values    eventValues
}

func NewEventFrame(eventType LogEventType, values eventValues, start, end int64, err error) *EventFrame {
	return &EventFrame{
		eventType: eventType,
		values:    values,
		baseFrame: baseFrame{
			frameType: LogFrameEvent,
			start:     start,
			end:       end,
			err:       err,
		},
	}
}

func (f EventFrame) Values() interface{} {
	return f.values
}

func (f EventFrame) EventType() LogEventType {
	return f.eventType
}

func (f EventFrame) String() string {
	var values []string
	for k, v := range f.Values().(map[string]interface{}) {
		values = append(values, fmt.Sprintf("%s: '%v'", k, v))
	}
	sort.Strings(values)

	return fmt.Sprintf("E frame: %s", strings.Join(values, ", "))
}

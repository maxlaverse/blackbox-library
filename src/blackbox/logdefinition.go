package blackbox

import (
	"github.com/pkg/errors"
)

type LogFrameType = byte

const (
	// See https://cleanflight.readthedocs.io/en/stable/development/Blackbox%20Internals/
	LogFrameEvent  LogFrameType = 69 // E
	LogFrameIntra               = 73 // I
	LogFrameInter               = 80 // P
	LogFrameSlow                = 83 // S
	LogFrameHeader              = 72 // H
	LogFrameGPS                 = 71 // G
)

var LogFrameAllTypes = []byte{
	LogFrameEvent,
	LogFrameIntra,
	LogFrameInter,
	LogFrameSlow,
	LogFrameHeader,
	LogFrameGPS,
}

// -------------------------------------------------------------------------- //

type LogEventType = byte

const (
	LogEventSyncBeep           LogEventType = 0
	LogEventInflightAdjustment              = 13
	LogEventLoggingResume                   = 14
	LogEventFlightMode                      = 30
	LogEventLogEnd                          = 255
)

// -------------------------------------------------------------------------- //

const (
	firmwareTypeUnknown = "Unknown firmware"
)

// LogDefinition represents a log
type LogDefinition struct {
	Product          string
	DataVersion      int
	LogStartDatetime string
	CraftName        string
	FieldsS          []FieldDefinition
	FieldsI          []FieldDefinition
	FieldsP          []FieldDefinition
	Headers          []Header
	Sysconfig        SysconfigType
	FieldIRL         map[FieldName]int
}

// FieldDefinition represents a field
type FieldDefinition struct {
	Name       FieldName
	Signed     bool
	Predictor  int64
	Encoding   int64
	GroupCount int
}

// SysconfigType represents a sysconfig
type SysconfigType struct {
	MinThrottle            int
	MaxThrottle            int
	MotorOutputLow         int
	MotorOutputHigh        int
	RcRate                 uint
	YawRate                uint
	Acc1G                  uint16
	GyroScale              float64
	Vbatscale              uint8
	Vbatmaxcellvoltage     uint8
	Vbatmincellvoltage     uint8
	Vbatwarningcellvoltage uint8
	CurrentMeterOffset     uint16
	CurrentMeterScale      uint16
	Vbatref                uint16
	FirmwareType           string
	FrameIntervalI         int
	FrameIntervalPNum      int
	FrameIntervalPDenom    int
}

func defaultLogDefinition() LogDefinition {
	return LogDefinition{
		Sysconfig: defaultSysconfig(),
	}
}

// defaultSysconfig returns a new Sysconfig with default value
func defaultSysconfig() SysconfigType {
	return SysconfigType{
		MinThrottle:            1150,
		MaxThrottle:            1850,
		MotorOutputLow:         1150,
		MotorOutputHigh:        1850,
		RcRate:                 90,
		YawRate:                0,
		Acc1G:                  1,
		GyroScale:              1,
		Vbatscale:              110,
		Vbatmaxcellvoltage:     43,
		Vbatmincellvoltage:     33,
		Vbatwarningcellvoltage: 35,
		CurrentMeterOffset:     0,
		CurrentMeterScale:      400,
		Vbatref:                4095,
		FirmwareType:           firmwareTypeUnknown,
		FrameIntervalI:         1,
		FrameIntervalPNum:      1,
		FrameIntervalPDenom:    1,
	}
}

// Header represents a header value
type Header struct {
	Name  HeaderName
	Value string
}

// GetHeaderValue returns the value of a header
func (f *LogDefinition) GetHeaderValue(headerName HeaderName) (string, error) {
	for _, v := range f.Headers {
		if v.Name == headerName {
			return v.Value, nil
		}
	}
	return "", errors.New("Not found")
}

// GetFieldIndex returns the position of a field
func (f *LogDefinition) GetFieldIndex(fieldName FieldName) (int, error) {
	index, ok := f.FieldIRL[fieldName]
	if !ok {
		return 0, errors.Errorf("Field definition for '%s' not found: %v", fieldName, f.FieldIRL)
	}
	return index, nil
}

package blackbox

import (
	"github.com/pkg/errors"
)

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
	FieldIRL         map[string]int
}

// FieldDefinition represents a field
type FieldDefinition struct {
	Name       string
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
}

// NewSysconfig returns a new Sysconfig with default value
func NewSysconfig() SysconfigType {
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
	}
}

// Header represents a header value
type Header struct {
	Name  string
	Value string
}

// GetHeaderValue returns the value of a header
func (f *LogDefinition) GetHeaderValue(headerName string) (string, error) {
	for _, v := range f.Headers {
		if v.Name == headerName {
			return v.Value, nil
		}
	}
	return "", errors.New("Not found")
}

// GetFieldIndex returns the position of a field
func (f *LogDefinition) GetFieldIndex(fieldName string) (int, error) {
	index, ok := f.FieldIRL[fieldName]
	if !ok {
		return 0, errors.Errorf("Field definition for '%s' not found: %v", fieldName, f.FieldIRL)
	}
	return index, nil
}

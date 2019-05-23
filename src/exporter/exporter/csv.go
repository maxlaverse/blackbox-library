package exporter

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/maxlaverse/blackbox-library/src/blackbox"
	"github.com/pkg/errors"
)

var fieldUnits = map[blackbox.FieldName]string{
	blackbox.FieldTime:             "us",
	blackbox.FieldVbatLatest:       "V",
	blackbox.FieldAmperageLatest:   "A",
	blackbox.FieldEnergyCumulative: "mAh",
	blackbox.FieldFlightModeFlags:  "flags",
	blackbox.FieldStateFlags:       "flags",
	blackbox.FieldFailsafePhase:    "flags",
}

// CsvFrameExporter transforms a FlightLog into a CSV file
type CsvFrameExporter struct {
	target         io.Writer
	lastSlowFrame  *blackbox.SlowFrame
	debugMode      bool
	hasAmperageAdc bool
	frameDef       blackbox.LogDefinition
	batteryState   batteryState
}

// NewCsvFrameExporter returns a new CsvFrameExporter
func NewCsvFrameExporter(file io.Writer, debugMode bool, frameDef blackbox.LogDefinition) *CsvFrameExporter {
	_, err := frameDef.GetFieldIndex(blackbox.FieldAmperageLatest)
	hasAmperageAdc := err == nil

	return &CsvFrameExporter{
		target:         file,
		lastSlowFrame:  blackbox.NewSlowFrame([]int32{0, 0, 0, 0, 0}, 0, 0),
		debugMode:      debugMode,
		frameDef:       frameDef,
		hasAmperageAdc: hasAmperageAdc,
		batteryState: batteryState{
			currentOffset: int32(frameDef.Sysconfig.CurrentMeterOffset),
			currentScale:  int32(frameDef.Sysconfig.CurrentMeterScale),
			vbatScale:     int32(frameDef.Sysconfig.Vbatscale),
		},
	}
}

// WriteHeaders writes the header line of the CSV file, with units
func (e *CsvFrameExporter) WriteHeaders() error {
	var headers []string
	for _, f := range e.frameDef.FieldsI {
		headers = append(headers, fieldWithUnit(f.Name))
	}
	if e.hasAmperageAdc {
		headers = append(headers, fieldWithUnit(blackbox.FieldEnergyCumulative))
	}
	for _, f := range e.frameDef.FieldsS {
		headers = append(headers, fieldWithUnit(f.Name))
	}
	return e.writeLn(strings.Join(headers, ", "))
}

// WriteFrame writes a frame line into the CSV with friendly values
func (e *CsvFrameExporter) WriteFrame(frame blackbox.Frame) error {
	switch frame.(type) {

	case *blackbox.EventFrame:
		if !e.debugMode {
			break
		}

		err := e.writeLn(frame.(*blackbox.EventFrame).String())
		if err != nil {
			return errors.Wrapf(err, "could not write frame '%s' to target file", string(frame.Type()))
		}

	case *blackbox.SlowFrame:
		e.lastSlowFrame = frame.(*blackbox.SlowFrame)
		if !e.debugMode {
			break
		}

		err := e.writeLn(frame.(*blackbox.SlowFrame).String())
		if err != nil {
			return errors.Wrapf(err, "could not write frame '%s' to target file", string(frame.Type()))
		}

	case *blackbox.MainFrame:
		values := e.friendlyMainFrameValues(frame.(*blackbox.MainFrame).Values().([]int32))

		err := e.writeLn(strings.Join(values, ", "))
		if err != nil {
			return errors.Wrapf(err, "could not write frame '%s' to target file", string(frame.Type()))
		}
	}
	return nil
}

func (e *CsvFrameExporter) friendlyMainFrameValues(valuesS []int32) []string {
	var values []string
	for k, v := range valuesS {
		if i, _ := e.frameDef.GetFieldIndex(blackbox.FieldVbatLatest); k == i {
			e.batteryState.setLatestVbat(v)
			values = append(values, prependSpaceForField(k, fmt.Sprintf("%.3f", math.Floor(e.batteryState.voltageVolt*1000)/1000)))
			continue
		}
		if i, _ := e.frameDef.GetFieldIndex(blackbox.FieldAmperageLatest); k == i {
			timeFieldIndex, _ := e.frameDef.GetFieldIndex(blackbox.FieldTime)
			e.batteryState.setLatestAmperage(v, valuesS[timeFieldIndex])
			values = append(values, prependSpaceForField(k, fmt.Sprintf("%.3f", e.batteryState.currentAmps)))
			continue
		}
		values = append(values, prependSpaceForField(k, fmt.Sprintf("%d", v)))
	}

	if e.hasAmperageAdc {
		values = append(values, prependSpaceForField(0, fmt.Sprintf("%f", e.batteryState.energyMilliampHours)))
	}

	for _, v := range e.lastSlowFrame.StringValues() {
		values = append(values, v)
	}
	return values
}

func (e *CsvFrameExporter) writeBytes(data []byte) error {
	_, err := e.target.Write(data)
	return err
}

func (e *CsvFrameExporter) writeLn(data string) error {
	return e.writeBytes(append([]byte(data), 10))
}

func fieldWithUnit(name blackbox.FieldName) string {
	v, ok := fieldUnits[name]
	if !ok {
		return string(name)
	}
	return fmt.Sprintf("%s (%s)", name, v)
}

func prependField(index int, name string) string {
	if index-len(name) < 0 {
		return name
	}
	return fmt.Sprintf("%s%s", strings.Repeat(" ", index-len(name)), name)
}

func prependSpaceForField(index int, name string) string {
	switch index {
	case 0, 1:
		return prependField(0, name)
	case 21, 22:
		return prependField(6, name)
	default:
		return prependField(3, name)
	}
}

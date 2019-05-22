package exporter

import (
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/maxlaverse/blackbox-library/src/blackbox"
	"github.com/pkg/errors"
)

const (
	adcVref = 33.0
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

type CsvFrameExporter struct {
	target             io.Writer
	lastSlow           *blackbox.SlowFrame
	NumberOfFramesRead int
	debugMode          bool
	frameDef           blackbox.LogDefinition
	state              currentMeterState
}

type currentMeterState struct {
	currentMilliamps    float64
	lastTime            int32
	energyMilliampHours float64
	hasAmperageAdc      bool
}

func NewCsvFrameExporter(file io.Writer, debugMode bool, frameDef blackbox.LogDefinition) *CsvFrameExporter {
	_, err := frameDef.GetFieldIndex(blackbox.FieldAmperageLatest)
	hasAmperageAdc := err == nil

	return &CsvFrameExporter{
		target:    file,
		lastSlow:  blackbox.NewSlowFrame([]int32{0, 0, 0, 0, 0}, 0, 0),
		debugMode: debugMode,
		frameDef:  frameDef,
		state: currentMeterState{
			hasAmperageAdc: hasAmperageAdc,
		},
	}
}

func (e *CsvFrameExporter) WriteHeaders() error {
	var headers []string
	for _, f := range e.frameDef.FieldsI {
		headers = append(headers, fieldWithUnit(f.Name))
	}
	if e.state.hasAmperageAdc {
		headers = append(headers, fieldWithUnit(blackbox.FieldEnergyCumulative))
	}
	for _, f := range e.frameDef.FieldsS {
		headers = append(headers, fieldWithUnit(f.Name))
	}
	return e.writeLn(strings.Join(headers, ", "))
}

func (e *CsvFrameExporter) WriteFrame(frame blackbox.Frame) error {
	e.NumberOfFramesRead++

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
		e.lastSlow = frame.(*blackbox.SlowFrame)
		if !e.debugMode {
			break
		}

		err := e.writeLn(frame.(*blackbox.SlowFrame).String())
		if err != nil {
			return errors.Wrapf(err, "could not write frame '%s' to target file", string(frame.Type()))
		}

	case *blackbox.MainFrame:
		values := e.computeMainFrameValues(frame.(*blackbox.MainFrame).Values().([]int32))
		err := e.writeLn(strings.Join(values, ", "))
		if err != nil {
			return errors.Wrapf(err, "could not write frame '%s' to target file", string(frame.Type()))
		}
	}
	return nil
}

func (e *CsvFrameExporter) writeLn(data string) error {
	return e.writeBytes(append([]byte(data), 10))
}

func (e *CsvFrameExporter) writeBytes(data []byte) error {
	_, err := e.target.Write(data)
	return err
}

func (e *CsvFrameExporter) computeMainFrameValues(valuesS []int32) []string {
	var values []string
	for k, v := range valuesS {
		if i, _ := e.frameDef.GetFieldIndex(blackbox.FieldVbatLatest); k == i {
			values = append(values, prependSpaceForField(k, fmt.Sprintf("%.3f", calculateVbatLatest(v, e.frameDef.Sysconfig.Vbatscale))))
			continue
		}
		if i, _ := e.frameDef.GetFieldIndex(blackbox.FieldAmperageLatest); k == i {
			e.state.currentMilliamps = calculateAmperageLatest(v, e.frameDef.Sysconfig.CurrentMeterOffset, e.frameDef.Sysconfig.CurrentMeterScale)
			values = append(values, prependSpaceForField(k, fmt.Sprintf("%.3f", e.state.currentMilliamps/1000)))
			continue
		}
		values = append(values, prependSpaceForField(k, fmt.Sprintf("%d", v)))
	}

	if e.state.hasAmperageAdc {
		if e.state.lastTime != 0.0 {
			e.state.energyMilliampHours += (e.state.currentMilliamps * float64(valuesS[1]-e.state.lastTime)) / (time.Hour.Seconds() * 1000000)
		}
		e.state.lastTime = valuesS[1]

		values = append(values, prependSpaceForField(0, fmt.Sprintf("%f", e.state.energyMilliampHours)))
	}

	for _, v := range e.lastSlow.StringValues() {
		values = append(values, v)
	}
	return values
}

func calculateVbatLatest(value int32, scale uint8) float64 {
	vbat := (float64(value) * adcVref * 10.0 * float64(scale)) / 4095.0
	return math.Floor(vbat) / 1000.0
}

func calculateAmperageLatest(value int32, offset, scale uint16) float64 {
	millivolts := float64((uint32(value)*adcVref*100)/4095) - float64(offset)
	millivolts = (millivolts * 10000) / float64(scale)
	return millivolts
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
	case 0:
		return prependField(0, name)
	case 1:
		return prependField(0, name)
	case 21:
		return prependField(6, name)
	case 22:
		return prependField(6, name)
	default:
		return prependField(3, name)
	}
}

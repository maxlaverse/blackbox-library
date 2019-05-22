package exporter

import (
	"fmt"
	"io"
	"math"
	"strings"

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
}

func NewCsvFrameExporter(file io.Writer, debugMode bool, frameDef blackbox.LogDefinition) *CsvFrameExporter {
	return &CsvFrameExporter{
		target:    file,
		lastSlow:  blackbox.NewSlowFrame([]int32{0, 0, 0, 0, 0}, 0, 0),
		debugMode: debugMode,
		frameDef:  frameDef,
	}
}

func (e *CsvFrameExporter) WriteHeaders() error {
	var headers []string
	for _, f := range e.frameDef.FieldsI {
		headers = append(headers, unitForField(f.Name))
	}
	for _, f := range e.frameDef.FieldsS {
		headers = append(headers, unitForField(f.Name))
	}

	_, err := e.target.Write([]byte(strings.Join(headers, ", ")))
	if err != nil {
		return errors.Wrap(err, "could not write headers to the target file")
	}

	return e.writeNewLine()
}

func (e *CsvFrameExporter) WriteFrame(frame blackbox.Frame) error {
	switch frame.(type) {

	case *blackbox.EventFrame:
		if !e.debugMode {
			return nil
		}

		_, err := e.target.Write([]byte(frame.(*blackbox.EventFrame).String()))
		if err != nil {
			return errors.Wrapf(err, "could not write frame '%s' to target file", string(frame.Type()))
		}

	case *blackbox.SlowFrame:
		e.NumberOfFramesRead++
		e.lastSlow = frame.(*blackbox.SlowFrame)
		if !e.debugMode {
			return nil
		}

		_, err := e.target.Write([]byte(frame.(*blackbox.SlowFrame).String()))
		if err != nil {
			return errors.Wrapf(err, "could not write frame '%s' to target file", string(frame.Type()))
		}

	case *blackbox.MainFrame:
		e.NumberOfFramesRead++
		var values []string
		for k, v := range frame.Values().([]int32) {
			if i, _ := e.frameDef.GetFieldIndex(blackbox.FieldVbatLatest); k == i {
				vbat := (float64(v) * adcVref * 10.0 * float64(e.frameDef.Sysconfig.Vbatscale)) / 4095.0
				values = append(values, prependSpaceForField(k, fmt.Sprintf("%.3f", math.Floor(vbat)/1000.0)))
				continue
			}
			if i, _ := e.frameDef.GetFieldIndex(blackbox.FieldAmperageLatest); k == i {
				millivolts := float64((uint32(v)*adcVref*100)/4095) - float64(e.frameDef.Sysconfig.CurrentMeterOffset)
				millivolts = (millivolts * 10000) / float64(e.frameDef.Sysconfig.CurrentMeterScale)
				values = append(values, prependSpaceForField(k, fmt.Sprintf("%.3f", millivolts/1000)))
				continue
			}
			values = append(values, prependSpaceForField(k, fmt.Sprintf("%d", v)))
		}
		for _, v := range e.lastSlow.StringValues() {
			values = append(values, v)
		}
		_, err := e.target.Write([]byte(strings.Join(values, ", ")))
		if err != nil {
			return errors.Wrapf(err, "could not write frame '%s' to target file", string(frame.Type()))
		}
	}
	return e.writeNewLine()
}

func (e *CsvFrameExporter) writeNewLine() error {
	_, err := e.target.Write([]byte("\n"))
	return errors.Wrap(err, "could not write to the target file")
}

func unitForField(name blackbox.FieldName) string {
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

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/maxlaverse/blackbox-library/src/blackbox"
)

func writeToCsv(f *blackbox.FlightLogReader, file *os.File) error {
	headers := []string{}
	for _, f := range f.FrameDef.FieldsI {
		headers = append(headers, unitForField(f.Name))
	}
	for _, f := range f.FrameDef.FieldsS {
		headers = append(headers, unitForField(f.Name))
	}

	_, err := file.WriteString(strings.Join(headers, ", "))
	if err != nil {
		return err
	}

	file.WriteString("\n")
	numberOfFramesRead := 0
	lastSlow := blackbox.NewFrame("S", []int32{0, 0, 0, 0, 0}, 0, 0)
	for _, frame := range f.Frames {
		if frame.Type == "E" {
			continue
		}
		numberOfFramesRead++
		if frame.Type == "S" {
			*lastSlow = frame
			continue
		}
		if frame.Type == "I" || frame.Type == "P" {
			values := []string{}
			for k, v := range frame.Values {
				//TODO: Read those header index dynamically
				if k == 21 {
					values = append(values, prependSpace(k, fmt.Sprintf("%.3f", (33.0*float64(v)*110)/4095.0/100.0)))
					continue
				}
				if k == 22 {
					millivolts := (uint32(v) * 33.0 * 100) / 4095
					millivolts -= 0

					//TODO: Implement millivolts reading and currentMeterScale
					values = append(values, prependSpace(k, fmt.Sprintf("%.3f", float64(int64(millivolts)*10000/400)/1000)))
					continue
				}
				values = append(values, prependSpace(k, fmt.Sprintf("%d", v)))
			}
			for k, v := range lastSlow.Values {
				values = append(values, parseFlags(k, fmt.Sprintf("%d", v)))
			}
			_, err := file.WriteString(strings.Join(values, ", "))
			if err != nil {
				return err
			}
		}
		file.WriteString("\n")
	}
	fmt.Printf("Wrote %d frames\n", numberOfFramesRead)
	return nil
}

func unitForField(name string) string {
	switch name {
	case "time":
		return fmt.Sprintf("%s (us)", name)
	case "vbatLatest":
		return fmt.Sprintf("%s (V)", name)
	case "amperageLatest":
		return fmt.Sprintf("%s (A)", name)
	case "energyCumulative":
		return fmt.Sprintf("%s (mAh)", name)
	case "flightModeFlags":
		return fmt.Sprintf("%s (flags)", name)
	case "stateFlags":
		return fmt.Sprintf("%s (flags)", name)
	case "failsafePhase":
		return fmt.Sprintf("%s (flags)", name)
	default:
		return name
	}
}

func fillTO(index int, name string) string {
	if index-len(name) < 0 {
		return name
	}
	return fmt.Sprintf("%s%s", strings.Repeat(" ", index-len(name)), name)
}
func prependSpace(index int, name string) string {
	switch index {
	case 0:
		return fillTO(0, name)
	case 1:
		return fillTO(0, name)
	case 21:
		return fillTO(6, name)
	case 22:
		return fillTO(6, name)
	default:
		return fillTO(3, name)
	}
}
func parseFlags(index int, name string) string {
	switch index {
	case 0:
		if name == "524289" {
			// 	"ANGLE_MODE",
			// 	"HORIZON_MODE",
			// 	"MAG",
			// 	"BARO",
			// 	"GPS_HOME",
			// 	"GPS_HOLD",
			// 	"HEADFREE",
			// 	"AUTOTUNE",
			// 	"PASSTHRU",
			// "SONAR"
			return "ANGLE_MODE"
		}
		return name
	case 1:
		if name == "8" {
			// 	"GPS_FIX_HOME",
			// 	"GPS_FIX",
			// 	"CALIBRATE_MAG",
			// 	"SMALL_ANGLE",
			// "FIXED_WING"
			return "SMALL_ANGLE"
		}
		return name
	case 2:
		if name == "0" {
			// 	"IDLE",
			// 	"RX_LOSS_DETECTED",
			// 	"LANDING",
			// "LANDED"
			return "IDLE"
		}
		return name
	case 3:
		return name
	default:
		return name
	}
}

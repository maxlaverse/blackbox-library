package blackbox

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
	"github.com/pkg/errors"
)

const (
	fieldProduct         = "Product"
	fieldDataVersion     = "Data version"
	fieldIName           = "Field I name"
	fieldISigned         = "Field I signed"
	fieldIPredictor      = "Field I predictor"
	fieldIEncoding       = "Field I encoding"
	fieldPPredictor      = "Field P predictor"
	fieldPEncoding       = "Field P encoding"
	fieldSName           = "Field S name"
	fieldSSigned         = "Field S signed"
	fieldSPredictor      = "Field S predictor"
	fieldSEncoding       = "Field S encoding"
	fieldVbatref         = "vbatref"
	fieldVbatcellvoltage = "vbatcellvoltage"
	fieldCurrentMeter    = "currentMeter"
	fieldMotorOutput     = "motorOutput"
	fieldFirmwareType    = "Firmware type"
)

// HeaderReader reads the headers of a log file
type HeaderReader struct {
	def LogDefinition
	enc *stream.Decoder
}

// NewHeaderReader returns a new HeaderReader
func NewHeaderReader(enc *stream.Decoder) HeaderReader {
	return HeaderReader{
		enc: enc,
		def: LogDefinition{
			Sysconfig: NewSysconfig(),
		},
	}
}

// ProcessHeaders process all the headers and returns the format of the data to come
func (h *HeaderReader) ProcessHeaders() (LogDefinition, error) {
	for {
		command, err := h.enc.NextByte()
		if err != nil {
			return h.def, err
		}
		if string(command) != "H" {
			return h.def, nil
		}

		var truc []string
		for {
			iteration, err := h.enc.ReadByte()
			if err != nil {
				return h.def, errors.WithStack(err)
			}
			if string(iteration) == "\n" {
				err := h.parseHeader(strings.Join(truc, ""))
				if err != nil {
					return h.def, err
				}
				break
			}

			truc = append(truc, string(iteration))
		}
	}
}

func (h *HeaderReader) parseHeader(out string) error {
	re := regexp.MustCompile(`H ([^:]+):(.*)`)
	match := re.FindStringSubmatch(out)

	switch match[1] {
	case fieldProduct:
		h.def.Product = match[2]

	case fieldFirmwareType:
		h.def.Sysconfig.FirmwareType = match[2]

	case fieldDataVersion:
		b, err := strconv.ParseInt(match[2], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse fieldDataVersion '%s' to int", match[2])
		}
		h.def.DataVersion = int(b)

	case fieldIName:
		fieldsRaw := strings.Split(match[2], ",")
		for _, fr := range fieldsRaw {
			d := FieldDefinition{
				Name: fr,
			}
			h.def.FieldsI = append(h.def.FieldsI, d)
			h.def.FieldsP = append(h.def.FieldsP, d)
		}

	case fieldISigned:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			b, err := strconv.ParseBool(fr)
			if err != nil {
				return errors.Errorf("Could not parse fieldISigned '%s' to bool", match[2])
			}
			h.def.FieldsI[i].Signed = b
		}

	case fieldIPredictor:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldIPredictor '%s' to int", match[2])
			}
			h.def.FieldsI[i].Predictor = n
		}

	case fieldIEncoding:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldIEncoding '%s' to int", match[2])
			}
			h.def.FieldsI[i].Encoding = n

		}

	case fieldPPredictor:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldPPredictor '%s' to int", match[2])
			}
			h.def.FieldsP[i].Predictor = n
		}

	case fieldPEncoding:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldPEncoding '%s' to int", match[2])
			}
			h.def.FieldsP[i].Encoding = n
		}

	case fieldSName:
		fieldsRaw := strings.Split(match[2], ",")
		for _, fr := range fieldsRaw {
			d := FieldDefinition{
				Name: fr,
			}
			h.def.FieldsS = append(h.def.FieldsS, d)
		}

	case fieldSSigned:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			b, err := strconv.ParseBool(fr)
			if err != nil {
				return errors.Errorf("Could not parse fieldSSigned '%s' to bool", match[2])
			}
			h.def.FieldsS[i].Signed = b
		}

	case fieldSPredictor:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldSPredictor '%s' to int", match[2])
			}
			h.def.FieldsS[i].Predictor = n
		}

	case fieldSEncoding:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldSEncoding '%s' to int", match[2])
			}
			h.def.FieldsS[i].Encoding = n
		}

	case fieldVbatref:
		val, err := strconv.ParseInt(match[2], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse fieldVbatref '%s' to int", match[2])
		}
		h.def.Sysconfig.Vbatref = uint16(val)

	case fieldVbatcellvoltage:
		header := Header{
			Name:  fieldVbatcellvoltage,
			Value: match[2],
		}
		h.def.Headers = append(h.def.Headers, header)

		vals := strings.Split(match[2], ",")
		val, err := strconv.ParseInt(vals[0], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse fieldVbatcellvoltage '%s' to int", vals[0])
		}
		h.def.Sysconfig.Vbatmincellvoltage = uint8(val)

		val, err = strconv.ParseInt(vals[1], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse fieldVbatcellvoltage '%s' to int", vals[1])
		}
		h.def.Sysconfig.Vbatwarningcellvoltage = uint8(val)

		val, err = strconv.ParseInt(vals[2], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse fieldVbatcellvoltage '%s' to int", vals[2])
		}
		h.def.Sysconfig.Vbatmaxcellvoltage = uint8(val)

	case fieldCurrentMeter:
		header := Header{
			Name:  fieldCurrentMeter,
			Value: match[2],
		}
		h.def.Headers = append(h.def.Headers, header)

		vals := strings.Split(match[2], ",")
		val, err := strconv.ParseInt(vals[0], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse fieldCurrentMeter '%s' to int", vals[0])
		}
		h.def.Sysconfig.CurrentMeterOffset = uint16(val)

		val, err = strconv.ParseInt(vals[1], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse fieldCurrentMeter '%s' to int", vals[1])
		}
		h.def.Sysconfig.CurrentMeterScale = uint16(val)

	case fieldMotorOutput:
		header := Header{
			Name:  fieldMotorOutput,
			Value: match[2],
		}
		h.def.Headers = append(h.def.Headers, header)

		vals := strings.Split(match[2], ",")
		val, err := strconv.ParseInt(vals[0], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse fieldMotorOutput '%s' to int", vals[0])
		}
		h.def.Sysconfig.MotorOutputLow = int(val)

		val, err = strconv.ParseInt(vals[1], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse fieldMotorOutput '%s' to int", vals[1])
		}
		h.def.Sysconfig.MotorOutputHigh = int(val)

	default:
		header := Header{
			Name:  match[1],
			Value: match[2],
		}
		h.def.Headers = append(h.def.Headers, header)
	}

	for i, field := range h.def.FieldsI {
		if field.Encoding == EncodingTag8_8SVB {
			groupCount := 0
			for j := i + 1; j < i+8 && j < len(h.def.FieldsI); j++ {
				groupCount = j - i
				if h.def.FieldsI[j].Encoding != EncodingTag8_8SVB {
					break
				}
			}
			for j := i; j < i+8 && j < len(h.def.FieldsI); j++ {
				h.def.FieldsI[j].GroupCount = groupCount
			}

		}
	}

	for i, field := range h.def.FieldsP {
		if field.Encoding == EncodingTag8_8SVB {
			groupCount := 0
			for j := i + 1; j < i+8 && j < len(h.def.FieldsP); j++ {
				groupCount = j - i
				if h.def.FieldsP[j].Encoding != EncodingTag8_8SVB {
					break
				}
			}
			for j := i; j < i+8 && j < len(h.def.FieldsP); j++ {
				h.def.FieldsP[j].GroupCount = groupCount
			}

		}
	}

	h.def.FieldIRL = map[string]int{}
	for i, field := range h.def.FieldsI {
		h.def.FieldIRL[field.Name] = i
	}

	return nil
}

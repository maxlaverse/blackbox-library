package blackbox

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
	"github.com/pkg/errors"
)

// HeaderName is the type for flight recorder headers
type HeaderName string

// List of all the headers used by the library
const (
	HeaderProduct         HeaderName = "Product"
	HeaderDataVersion     HeaderName = "Data version"
	HeaderIName           HeaderName = "Field I name"
	HeaderISigned         HeaderName = "Field I signed"
	HeaderIPredictor      HeaderName = "Field I predictor"
	HeaderIEncoding       HeaderName = "Field I encoding"
	HeaderPPredictor      HeaderName = "Field P predictor"
	HeaderPEncoding       HeaderName = "Field P encoding"
	HeaderSName           HeaderName = "Field S name"
	HeaderSSigned         HeaderName = "Field S signed"
	HeaderSPredictor      HeaderName = "Field S predictor"
	HeaderSEncoding       HeaderName = "Field S encoding"
	HeaderVbatref         HeaderName = "vbatref"
	HeaderVbatcellvoltage HeaderName = "vbatcellvoltage"
	HeaderCurrentMeter    HeaderName = "currentMeter"
	HeaderMotorOutput     HeaderName = "motorOutput"
	HeaderFirmwareType    HeaderName = "Firmware type"
	HeaderIInterval       HeaderName = "I interval"
	HeaderPInterval       HeaderName = "P interval"

	headerRegExp = `H ([^:]+):(.*)`
)

// HeaderReader reads the headers of a log file
type HeaderReader struct {
	def LogDefinition
	enc *stream.Decoder
	re  *regexp.Regexp
}

// NewHeaderReader returns a new HeaderReader
func NewHeaderReader(enc *stream.Decoder) HeaderReader {
	return HeaderReader{
		enc: enc,
		def: defaultLogDefinition(),
		re:  regexp.MustCompile(headerRegExp),
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

		var byteBuffer []string
		for {
			iteration, err := h.enc.ReadByte()
			if err != nil {
				return h.def, errors.WithStack(err)
			}
			if string(iteration) == "\n" {
				err := h.parseHeader(strings.Join(byteBuffer, ""))
				if err != nil {
					return h.def, err
				}
				break
			}
			byteBuffer = append(byteBuffer, string(iteration))
		}
	}
}

func (h *HeaderReader) parseHeader(out string) error {
	match := h.re.FindStringSubmatch(out)

	switch HeaderName(match[1]) {
	case HeaderProduct:
		h.def.Product = match[2]

	case HeaderFirmwareType:
		h.def.Sysconfig.FirmwareType = match[2]

	case HeaderDataVersion:
		b, err := strconv.ParseInt(match[2], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse fieldDataVersion '%s' to int", match[2])
		}
		h.def.DataVersion = int(b)

	case HeaderIName:
		fieldsRaw := strings.Split(match[2], ",")
		for _, fr := range fieldsRaw {
			d := FieldDefinition{
				Name: FieldName(fr),
			}
			h.def.FieldsI = append(h.def.FieldsI, d)
			h.def.FieldsP = append(h.def.FieldsP, d)
		}

	case HeaderISigned:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			b, err := strconv.ParseBool(fr)
			if err != nil {
				return errors.Errorf("Could not parse fieldISigned '%s' to bool", match[2])
			}
			h.def.FieldsI[i].Signed = b
		}

	case HeaderIPredictor:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldIPredictor '%s' to int", match[2])
			}
			h.def.FieldsI[i].Predictor = n
		}

	case HeaderIEncoding:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldIEncoding '%s' to int", match[2])
			}
			h.def.FieldsI[i].Encoding = n

		}

	case HeaderPPredictor:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldPPredictor '%s' to int", match[2])
			}
			h.def.FieldsP[i].Predictor = n
		}

	case HeaderPEncoding:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldPEncoding '%s' to int", match[2])
			}
			h.def.FieldsP[i].Encoding = n
		}

	case HeaderSName:
		fieldsRaw := strings.Split(match[2], ",")
		for _, fr := range fieldsRaw {
			d := FieldDefinition{
				Name: FieldName(fr),
			}
			h.def.FieldsS = append(h.def.FieldsS, d)
		}

	case HeaderSSigned:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			b, err := strconv.ParseBool(fr)
			if err != nil {
				return errors.Errorf("Could not parse fieldSSigned '%s' to bool", match[2])
			}
			h.def.FieldsS[i].Signed = b
		}

	case HeaderSPredictor:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldSPredictor '%s' to int", match[2])
			}
			h.def.FieldsS[i].Predictor = n
		}

	case HeaderSEncoding:
		fieldsRaw := strings.Split(match[2], ",")
		for i, fr := range fieldsRaw {
			n, err := strconv.ParseInt(fr, 10, 8)
			if err != nil {
				return errors.Errorf("Could not parse fieldSEncoding '%s' to int", match[2])
			}
			h.def.FieldsS[i].Encoding = n
		}

	case HeaderVbatref:
		val, err := strconv.ParseInt(match[2], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse fieldVbatref '%s' to int", match[2])
		}
		h.def.Sysconfig.Vbatref = uint16(val)

	case HeaderVbatcellvoltage:
		header := Header{
			Name:  HeaderVbatcellvoltage,
			Value: match[2],
		}
		h.def.Headers = append(h.def.Headers, header)

		vals := strings.Split(match[2], ",")
		val, err := strconv.ParseInt(vals[0], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse first part of fieldVbatcellvoltage '%s' to int", vals[0])
		}
		h.def.Sysconfig.Vbatmincellvoltage = uint8(val)

		val, err = strconv.ParseInt(vals[1], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse second part of fieldVbatcellvoltage '%s' to int", vals[1])
		}
		h.def.Sysconfig.Vbatwarningcellvoltage = uint8(val)

		val, err = strconv.ParseInt(vals[2], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse third part of fieldVbatcellvoltage '%s' to int", vals[2])
		}
		h.def.Sysconfig.Vbatmaxcellvoltage = uint8(val)

	case HeaderCurrentMeter:
		header := Header{
			Name:  HeaderCurrentMeter,
			Value: match[2],
		}
		h.def.Headers = append(h.def.Headers, header)

		vals := strings.Split(match[2], ",")
		val, err := strconv.ParseInt(vals[0], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse first part of fieldCurrentMeter '%s' to int", vals[0])
		}
		h.def.Sysconfig.CurrentMeterOffset = uint16(val)

		val, err = strconv.ParseInt(vals[1], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse second part of fieldCurrentMeter '%s' to int", vals[1])
		}
		h.def.Sysconfig.CurrentMeterScale = uint16(val)

	case HeaderMotorOutput:
		header := Header{
			Name:  HeaderMotorOutput,
			Value: match[2],
		}
		h.def.Headers = append(h.def.Headers, header)

		vals := strings.Split(match[2], ",")
		val, err := strconv.ParseInt(vals[0], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse first part of fieldMotorOutput '%s' to int", vals[0])
		}
		h.def.Sysconfig.MotorOutputLow = int(val)

		val, err = strconv.ParseInt(vals[1], 10, 32)
		if err != nil {
			return errors.Errorf("Could not parse second part of fieldMotorOutput '%s' to int", vals[1])
		}
		h.def.Sysconfig.MotorOutputHigh = int(val)
	case HeaderIInterval:
		frameIntervalI, err := strconv.ParseInt(match[2], 10, 32)
		if err != nil {
			panic(err)
		}
		if frameIntervalI < 1 {
			frameIntervalI = 1
		}
		h.def.Sysconfig.FrameIntervalI = int(frameIntervalI)

	case HeaderPInterval:
		parts := strings.Split(match[2], "/")

		if len(parts) > 1 {
			frameIntervalPNum, err := strconv.ParseInt(parts[0], 10, 32)
			if err != nil {
				panic(err)
			}
			h.def.Sysconfig.FrameIntervalPNum = int(frameIntervalPNum)

			frameIntervalPDeNum, err := strconv.ParseInt(parts[1], 10, 32)
			if err != nil {
				panic(err)
			}
			h.def.Sysconfig.FrameIntervalPDenom = int(frameIntervalPDeNum)
		}

	default:
		header := Header{
			Name:  HeaderName(match[1]),
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

	h.def.FieldIRL = map[FieldName]int{}
	for i, field := range h.def.FieldsI {
		h.def.FieldIRL[field.Name] = i
	}

	return nil
}

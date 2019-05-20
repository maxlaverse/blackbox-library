package blackbox

import (
	"bufio"
	"fmt"
	"io"

	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
)

const (
	defaultBufferSize = 8192
)

// FlightLogReader parses logs generated by the CleanFlight controller
type FlightLogReader struct {
	FrameDef LogDefinition
	Frames   []Frame
	opts     FlightLogReaderOpts
}

// FlightLogReaderOpts holds the options to get a new FlightLogReader
type FlightLogReaderOpts struct {
	Raw bool
}

// NewFlightLogReader returns a new FlightLogReader
func NewFlightLogReader(opts FlightLogReaderOpts) FlightLogReader {
	return FlightLogReader{
		opts: opts,
	}
}

// LoadFile reads a file
func (f *FlightLogReader) LoadFile(file io.Reader) error {
	bufferedStream := bufio.NewReaderSize(file, defaultBufferSize)

	decoder := stream.NewDecoder(bufferedStream)

	headerReader := NewHeaderReader(&decoder)
	frameDefinition, err := headerReader.ProcessHeaders()
	if err != nil {
		return err
	}

	f.FrameDef = frameDefinition

	opts := &FrameReaderOptions{
		Raw: false,
	}

	frameReader := NewFrameReader(&decoder, frameDefinition, opts)
	for !frameReader.Finished {
		frame, err := frameReader.ReadNextFrame()
		if err != nil {
			fmt.Printf("Frame was discarded: %v\n", err)
			consumeToNext(decoder)
			continue
		}

		f.Frames = append(f.Frames, *frame)
	}

	frameReader.PrintStatistics()
	return nil
}

func consumeToNext(enc stream.Decoder) int64 {
	defer fmt.Printf("\n")
	intialPos := enc.BytesRead()
	for i := 0; i < 256; i++ {
		b, err := enc.NextByte()
		if err != nil {
			panic(err)
		}

		fmt.Printf("%d:%v, ", i, string(b))
		if string(b) == "S" || string(b) == "H" || string(b) == "E" || string(b) == "I" || string(b) == "P" || string(b) == "G" {
			newPos := enc.BytesRead()
			return newPos - intialPos
		}
		enc.ReadByte()
	}
	panic("Could not find next frame")
}

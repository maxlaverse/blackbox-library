package blackbox

import (
	"bufio"
	"context"
	"io"

	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
)

const (
	defaultBufferSize = 8192
)

// FlightLogReader parses logs generated by the CleanFlight controller
type FlightLogReader struct {
	FrameDef LogDefinition
	Stats    *StatsType
	opts     FlightLogReaderOpts
}

// FlightLogReaderOpts holds the options to get a new FlightLogReader
type FlightLogReaderOpts struct {
	Raw bool
}

// NewFlightLogReader returns a new FlightLogReader
func NewFlightLogReader(opts FlightLogReaderOpts) *FlightLogReader {
	return &FlightLogReader{
		opts: opts,
	}
}

// LoadFile reads flight logs from a file.
// Accepts context and stops processing when the context is canceled.
// Returns channel with successfully parsed frames and channel with errors.
func (f *FlightLogReader) LoadFile(file io.Reader, ctx context.Context) (<-chan Frame, error) {
	frameReader, err := f.initFrameReader(file)
	if err != nil {
		return nil, err
	}

	frameChan := make(chan Frame)
	go func() {
		defer func() {
			close(frameChan)
		}()

		// collect stats when the process is done
		defer func() {
			f.Stats = frameReader.Stats
		}()

		for !frameReader.Finished {
			// check on every iteration if the context has been canceled and stop processing if so
			select {
			case <-ctx.Done():
				return
			default:
			}

			frame := frameReader.ReadNextFrame()

			if frame.Error() != nil {
				// Try to consume the remaining bytes from the broken frame
				if _, err = frameReader.ConsumeToNext(); err != nil {
					frame.setError(err)
					frameChan <- frame
					return
				}
			}

			// finally send the frame out
			frameChan <- frame
		}
	}()

	return frameChan, nil
}

func (f *FlightLogReader) initFrameReader(file io.Reader) (*FrameReader, error) {
	bufferedStream := bufio.NewReaderSize(file, defaultBufferSize)

	decoder := stream.NewDecoder(bufferedStream)
	headerReader := NewHeaderReader(decoder)
	frameDefinition, err := headerReader.ProcessHeaders()
	if err != nil {
		return nil, err
	}

	f.FrameDef = frameDefinition

	opts := &FrameReaderOptions{
		Raw: f.opts.Raw,
	}
	return NewFrameReader(decoder, frameDefinition, opts)
}

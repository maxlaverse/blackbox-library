package blackbox

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkReadFrames(b *testing.B) {
	logFile, err := os.Open("../../fixtures/normal.bfl")
	assert.NoError(b, err)
	defer logFile.Close()

	for i := 0; i < b.N; i++ {
		logFile.Seek(0, io.SeekStart)
		frameRead := 0
		bytesRead := 0

		flightLog := NewFlightLogReader(FlightLogReaderOpts{Raw: true})
		frameChan, errChan, err := flightLog.LoadFile(logFile, context.Background())
		assert.NoError(b, err)

	Loop:
		for {
			select {
			case frame := <-frameChan:
				if frame == nil {
					break Loop
				}
				bytesRead = bytesRead + frame.Size()
				frameRead++
			case <-errChan:
				break Loop
			}
		}
		assert.Equal(b, 10, frameRead)
		assert.Equal(b, 209, bytesRead)
	}
}

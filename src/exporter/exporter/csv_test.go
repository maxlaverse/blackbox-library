package exporter

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/maxlaverse/blackbox-library/src/blackbox"
	"github.com/stretchr/testify/assert"
)

func TestCsvHeaders(t *testing.T) {
	var csvBuffer bytes.Buffer
	csvExporter := NewCsvFrameExporter(&csvBuffer, true)

	flightLog, frameChan, errChan, logFile := readFixture(t, "normal.bfl", blackbox.FlightLogReaderOpts{Raw: true})
	defer logFile.Close()

	select {
	case <-frameChan:
		err := csvExporter.WriteHeaders(flightLog.FrameDef)
		assert.NoError(t, err)
	case err := <-errChan:
		assert.NoError(t, err)
	}

	assert.Equal(t, "loopIteration, time (us), axisP[0], axisP[1], axisP[2], axisI[0], axisI[1], axisI[2], axisD[0], axisD[1], axisF[0], axisF[1], axisF[2], rcCommand[0], rcCommand[1], rcCommand[2], rcCommand[3], setpoint[0], setpoint[1], setpoint[2], setpoint[3], vbatLatest (V), amperageLatest (A), rssi, gyroADC[0], gyroADC[1], gyroADC[2], accSmooth[0], accSmooth[1], accSmooth[2], debug[0], debug[1], debug[2], debug[3], motor[0], motor[1], motor[2], motor[3], flightModeFlags (flags), stateFlags (flags), failsafePhase (flags), rxSignalReceived, rxFlightChannelsValid\n", csvBuffer.String())
}

func TestRawValues(t *testing.T) {
	expectedLines := []string{
		"E frame: currentTime: 55158008, iteration: 52992\n",
		"52992, 55158008,  -1,  -4,  -1,   5,  -2,  -1,  -1, -30,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, -2.287, 15.200, 785,   0,   3,   3,  60,   8, 2232,  -5, -14,   1,   0, 333, 129,  -2, 144, 0, 0, IDLE, 0, 0\n",
		"E frame: beepTime: 41780625\n",
		"E frame: flags: 524289, lastFlags: 1\n",
		"S frame: ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"0, -2147483150,   1,   0,   1,   0,   0,   0,  -4,  -5,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,  0.000,  0.000,   0,   0,   1,  -1,   0,  -2,  -1,   2,   5,  -2,   0,  -5,  16, -14,   4, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"0, -2147483648,   0,   1,  -1,   0,   0,   0,  -1,   4,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,  0.000,  0.000,   0,  -1,   0,   1,   0,  -1,   0,   3,   9,   0,   0,   7,   0,   0,  -6, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"0, -2147483645,   2,   3,   0,   0,   0,   0,   4,  12,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,  0.000,  0.000,   0,  -3,  -2,   1,   0,  -2,  -2,   1,   9,   1,   0,  24, -46,  43, -20, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"0, 2147483641,  -1,   3,   0,   0,   0,   0,   7,  20,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,  0.000,  0.000,   0,   0,  -4,   0,   0,  -1,  -1,   1,   9,   2,   0,  42, -79,  80, -41, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"E frame: data: [69 110 100 32 111 102 32 108 111 103 0 10]\n",
	}

	var csvBuffer bytes.Buffer
	csvExporter := NewCsvFrameExporter(&csvBuffer, true)

	_, frameChan, errChan, logFile := readFixture(t, "normal.bfl", blackbox.FlightLogReaderOpts{Raw: true})
	defer logFile.Close()
	for k, line := range expectedLines {
		t.Run(fmt.Sprintf("Line %d", k), func(t *testing.T) {
			select {
			case frame := <-frameChan:
				err := csvExporter.WriteFrame(frame)
				assert.NoError(t, err)
			case err := <-errChan:
				assert.NoError(t, err)
			}
			assert.Equal(t, line, csvBuffer.String())
		})
		csvBuffer.Reset()
	}
}

func TestNormalValues(t *testing.T) {
	expectedLines := []string{
		"E frame: currentTime: 55158008, iteration: 52992\n",
		"52992, 55158008,  -1,  -4,  -1,   5,  -2,  -1,  -1, -30,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,   0,   3,   3,  60,   8, 2232,  -5, -14,   1,   0, 521, 650, 519, 665, 0, 0, IDLE, 0, 0\n",
		"E frame: beepTime: 41780625\n",
		"E frame: flags: 524289, lastFlags: 1\n",
		"S frame: ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"52993, 55158507,   0,  -4,   0,   5,  -2,  -1,  -5, -35,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,   0,   4,   2,  60,   6, 2231,  -3,  -9,  -1,   0, 516, 666, 505, 669, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"52994, 55159007,   0,  -3,  -1,   5,  -2,  -1,  -6, -31,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,  -1,   3,   3,  60,   6, 2231,  -1,  -2,   0,   0, 525, 658, 512, 661, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"52995, 55159511,   2,   0,  -1,   5,  -2,  -1,  -2, -19,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,  -3,   1,   3,  60,   4, 2229,  -1,   4,   1,   0, 544, 616, 551, 645, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"52996, 55160009,   1,   3,  -1,   5,  -2,  -1,   5,   1,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,  -2,  -2,   3,  60,   4, 2229,   0,  10,   2,   0, 576, 558, 611, 612, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"E frame: data: [69 110 100 32 111 102 32 108 111 103 0 10]\n",
	}

	var csvBuffer bytes.Buffer
	csvExporter := NewCsvFrameExporter(&csvBuffer, true)

	_, frameChan, errChan, logFile := readFixture(t, "normal.bfl", blackbox.FlightLogReaderOpts{Raw: false})
	defer logFile.Close()
	for k, line := range expectedLines {
		t.Run(fmt.Sprintf("Line %d", k), func(t *testing.T) {
			select {
			case frame := <-frameChan:
				err := csvExporter.WriteFrame(frame)
				assert.NoError(t, err)
			case err := <-errChan:
				assert.NoError(t, err)
			}
			assert.Equal(t, line, csvBuffer.String())
		})
		csvBuffer.Reset()
	}
}

func readFixture(t *testing.T, fixtureFile string, opts blackbox.FlightLogReaderOpts) (*blackbox.FlightLogReader, <-chan blackbox.Frame, <-chan error, *os.File) {
	flightLog := blackbox.NewFlightLogReader(opts)
	logFile, err := os.Open(fmt.Sprintf("../../../fixtures/%s", fixtureFile))
	assert.NoError(t, err)

	frameChan, errChan := flightLog.LoadFile(logFile, context.Background())
	return flightLog, frameChan, errChan, logFile
}
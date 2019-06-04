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
	frameDef, frameChan, logFile := readFixture(t, "normal.bfl", blackbox.FlightLogReaderOpts{Raw: true})
	defer logFile.Close()

	var csvBuffer bytes.Buffer
	csvExporter := NewCsvFrameExporter(&csvBuffer, true, frameDef)

	select {
	case <-frameChan:
		err := csvExporter.WriteHeaders()
		assert.NoError(t, err)
	}

	assert.Equal(t, "loopIteration, time (us), axisP[0], axisP[1], axisP[2], axisI[0], axisI[1], axisI[2], axisD[0], axisD[1], axisF[0], axisF[1], axisF[2], rcCommand[0], rcCommand[1], rcCommand[2], rcCommand[3], setpoint[0], setpoint[1], setpoint[2], setpoint[3], vbatLatest (V), amperageLatest (A), rssi, gyroADC[0], gyroADC[1], gyroADC[2], accSmooth[0], accSmooth[1], accSmooth[2], debug[0], debug[1], debug[2], debug[3], motor[0], motor[1], motor[2], motor[3], energyCumulative (mAh), flightModeFlags (flags), stateFlags (flags), failsafePhase (flags), rxSignalReceived, rxFlightChannelsValid\n", csvBuffer.String())
}

func TestRawValues(t *testing.T) {
	expectedLines := []string{
		"E frame: currentTime: '55158008', iteration: '52992', name: 'Logging resume'\n",
		"52992, 55158008,  -1,  -4,  -1,   5,  -2,  -1,  -1, -30,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, -2.288, 15.200, 785,   0,   3,   3,  60,   8, 2232,  -5, -14,   1,   0, 333, 129,  -2, 144, 0.000000, 0, 0, IDLE, 0, 0, I, offset: 1573, size: 53\n",
		"E frame: beepTime: '41780625', name: 'Sync beep'\n",
		"E frame: flags: '524289', lastFlags: '1', name: 'Flight mode'\n",
		"S frame: ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"52993, 499,   1,   0,   1,   0,   0,   0,  -4,  -5,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,  0.000,  0.000,   0,   0,   1,  -1,   0,  -2,  -1,   2,   5,  -2,   0,  -5,  16, -14,   4, 0.000000, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1, P, offset: 1644, size: 29\n",
		"52994, 1,   0,   1,  -1,   0,   0,   0,  -1,   4,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,  0.000,  0.000,   0,  -1,   0,   1,   0,  -1,   0,   3,   9,   0,   0,   7,   0,   0,  -6, 0.000000, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1, P, offset: 1673, size: 28\n",
		"52995, 4,   2,   3,   0,   0,   0,   0,   4,  12,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,  0.000,  0.000,   0,  -3,  -2,   1,   0,  -2,  -2,   1,   9,   1,   0,  24, -46,  43, -20, 0.000000, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1, P, offset: 1701, size: 28\n",
		"52996, -6,  -1,   3,   0,   0,   0,   0,   7,  20,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,   0,  0.000,  0.000,   0,   0,  -4,   0,   0,  -1,  -1,   1,   9,   2,   0,  42, -79,  80, -41, 0.000000, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1, P, offset: 1729, size: 30\n",
		"E frame: data: '[69 110 100 32 111 102 32 108 111 103 0 10]', name: 'Log clean end'\n",
	}

	frameDef, frameChan, logFile := readFixture(t, "normal.bfl", blackbox.FlightLogReaderOpts{Raw: true})
	defer logFile.Close()

	var csvBuffer bytes.Buffer
	csvExporter := NewCsvFrameExporter(&csvBuffer, true, frameDef)

	for k, line := range expectedLines {
		t.Run(fmt.Sprintf("Line %d", k), func(t *testing.T) {
			frame := <-frameChan
			assert.NoError(t, frame.Error())
			err := csvExporter.WriteFrame(frame)
			assert.NoError(t, err)
			assert.Equal(t, line, csvBuffer.String())
		})
		csvBuffer.Reset()
	}
}

func TestNormalValues(t *testing.T) {
	expectedLines := []string{
		"E frame: currentTime: '55158008', iteration: '52992', name: 'Logging resume'\n",
		"52992, 55158008,  -1,  -4,  -1,   5,  -2,  -1,  -1, -30,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,   0,   3,   3,  60,   8, 2232,  -5, -14,   1,   0, 521, 650, 519, 665, 0.000000, 0, 0, IDLE, 0, 0, I, offset: 1573, size: 53\n",
		"E frame: beepTime: '41780625', name: 'Sync beep'\n",
		"E frame: flags: '524289', lastFlags: '1', name: 'Flight mode'\n",
		"S frame: ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1\n",
		"52993, 55158507,   0,  -4,   0,   5,  -2,  -1,  -5, -35,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,   0,   4,   2,  60,   6, 2231,  -3,  -9,  -1,   0, 516, 666, 505, 669, 0.002107, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1, P, offset: 1644, size: 29\n",
		"52994, 55159007,   0,  -3,  -1,   5,  -2,  -1,  -6, -31,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,  -1,   3,   3,  60,   6, 2231,  -1,  -2,   0,   0, 525, 658, 512, 661, 0.004218, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1, P, offset: 1673, size: 28\n",
		"52995, 55159511,   2,   0,  -1,   5,  -2,  -1,  -2, -19,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,  -3,   1,   3,  60,   4, 2229,  -1,   4,   1,   0, 544, 616, 551, 645, 0.006346, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1, P, offset: 1701, size: 28\n",
		"52996, 55160009,   1,   3,  -1,   5,  -2,  -1,   5,   1,   0,   0,   0,  -1,   0,   4, 1216,  -1,   0,   2, 216, 14.245, 15.200, 785,  -2,  -2,   3,  60,   4, 2229,   0,  10,   2,   0, 576, 558, 611, 612, 0.008449, ANGLE_MODE, SMALL_ANGLE, IDLE, 1, 1, P, offset: 1729, size: 30\n",
		"E frame: data: '[69 110 100 32 111 102 32 108 111 103 0 10]', name: 'Log clean end'\n",
	}

	frameDef, frameChan, logFile := readFixture(t, "normal.bfl", blackbox.FlightLogReaderOpts{Raw: false})
	defer logFile.Close()

	var csvBuffer bytes.Buffer
	csvExporter := NewCsvFrameExporter(&csvBuffer, true, frameDef)

	for k, line := range expectedLines {
		t.Run(fmt.Sprintf("Line %d", k), func(t *testing.T) {
			frame := <-frameChan
			assert.NoError(t, frame.Error())
			err := csvExporter.WriteFrame(frame)
			assert.NoError(t, err)
			assert.Equal(t, line, csvBuffer.String())
		})
		csvBuffer.Reset()
	}
}

func readFixture(t *testing.T, fixtureFile string, opts blackbox.FlightLogReaderOpts) (blackbox.LogDefinition, <-chan blackbox.Frame, *os.File) {
	flightLog := blackbox.NewFlightLogReader(opts)
	logFile, err := os.Open(fmt.Sprintf("../../../fixtures/%s", fixtureFile))
	assert.NoError(t, err)

	frameChan, err := flightLog.LoadFile(context.Background(), logFile)
	assert.NoError(t, err)

	return flightLog.FrameDef, frameChan, logFile
}

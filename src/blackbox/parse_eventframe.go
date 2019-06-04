package blackbox

import (
	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
	"github.com/pkg/errors"
)

func parseEventFrame(dec *stream.Decoder) (LogEventType, eventValues, error) {
	values := make(eventValues)
	eventType, err := dec.ReadByte()
	if err != nil {
		return 0, nil, err
	}

	switch eventType {
	case LogEventSyncBeep:
		beepTime, err := dec.ReadUnsignedVB()
		if err != nil {
			return 0, nil, err
		}
		values["name"] = "Sync beep"
		values["beepTime"] = beepTime

	case LogEventInflightAdjustment:
		return 0, nil, errors.New("Not implemented: logEventInflightAdjustment")

	case LogEventLoggingResume:
		val, err := dec.ReadUnsignedVB()
		if err != nil {
			return 0, nil, err
		}
		values["iteration"] = int64(val)

		val, err = dec.ReadUnsignedVB()
		if err != nil {
			return 0, nil, err
		}
		values["name"] = "Logging resume"
		values["currentTime"] = int64(val)

	case LogEventFlightMode:
		flags, err := dec.ReadUnsignedVB()
		if err != nil {
			return 0, nil, err
		}
		lastFlags, err := dec.ReadUnsignedVB()
		if err != nil {
			return 0, nil, err
		}
		values["name"] = "Flight mode"
		values["flags"] = flags
		values["lastFlags"] = lastFlags

	case LogEventLogEnd:
		val, err := dec.ReadBytes(12)
		if err != nil {
			return 0, nil, err
		}

		reachedEndOfFile, err := dec.EOF()
		if err != nil {
			return 0, nil, err
		}
		if !reachedEndOfFile {
			return 0, nil, errors.New("There are additional data after the end of the file")
		}

		values["name"] = "Log clean end"
		values["data"] = val

	default:
		return 0, nil, errors.Errorf("Event type is unknown - ignored: %v\n", eventType)
	}

	return eventType, values, nil
}

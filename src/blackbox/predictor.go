package blackbox

import (
	"github.com/pkg/errors"
)

const (
	// PredictorZero returns the value unmodified
	PredictorZero = 0

	// PredictorPrevious return the value substracted from the interframe
	PredictorPrevious = 1

	// PredictorStraightLine assumes that the slope between the current measurement and the previous one will be similar to the slope between the previous measurement and the one before that.
	// This is common for fields which increase at a steady rate, such as the "time" field. The predictor is `history_age_2 - 2 * history_age_1`.
	PredictorStraightLine = 2

	// PredicatorAverage2 is the average of the two previously logged values of the field (i.e. `(history_age_1 + history_age_2) / 2`).
	// It is used when there is significant random noise involved in the field, which means that the average of the recent history is a better predictor of the next value than the previous value on its own would be (for example, in gyroscope or motor measurements).
	PredicatorAverage2 = 3

	// PredictorMinThrottle subtracts the value of "minthrottle" which is included in the log header.
	// In Cleanflight, motors always lie in the range of `[minthrottle ... maxthrottle]` when the craft is armed, so this predictor is used for the first motor value in intraframes.
	PredictorMinThrottle = 4

	// PredictorMotor0 is set to the value of `motor[0]` which was decoded earlier within the current frame.
	// It is used in intraframes for every motor after the first one, because the motor commands typically lie in a tight grouping.
	PredictorMotor0 = 5

	// PredictorInc assumes that the field will be incremented by 1 unit for every main loop iteration. This is used to predict the `loopIteration` field, which increases by 1 for every loop iteration.
	PredictorInc = 6

	// Predictor1500 is set to a fixed value of 1500.
	// It is preferred for logging servo values in intraframes, since these  typically lie close to the midpoint of 1500us.
	Predictor1500 = 8

	// PredictorVbatRef is set to the "vbatref" field written in the log header.
	// It is used when logging intraframe battery voltages in Cleanflight, since these are expected to be broadly similar to the first battery voltage seen during arming.
	PredictorVbatRef = 9

	// PredictorMinMotor returns the value and the minimum motor low output summed
	PredictorMinMotor = 11
)

// ApplyPrediction a predictor on a field and return the resulting value
func ApplyPrediction(frameDef LogDefinition, values []int32, fieldIndex int, predictor int, value int32, previous *MainFrame, previous2 *MainFrame) (int32, error) {

	// First see if we have a prediction that doesn't require a previous frame as reference:
	switch predictor {
	case PredictorZero:
		// No correction to apply
		break
	case PredictorMinThrottle:
		value += int32(frameDef.Sysconfig.MinThrottle)
	case Predictor1500:
		value += 1500
	case PredictorMotor0:
		motor0idx, err := frameDef.GetFieldIndex("motor[0]")
		if err != nil {
			return value, err
		}
		value += values[motor0idx]
	case PredictorVbatRef:
		value += int32(frameDef.Sysconfig.Vbatref)
	case PredictorPrevious:
		if previous == nil {
			break
		}
		value = value + previous.values[fieldIndex]
	case PredictorStraightLine:
		if previous == nil {
			break
		}
		value = value + 2*previous.values[fieldIndex] - previous2.values[fieldIndex]
	case PredicatorAverage2:
		if previous == nil {
			break
		}
		value = value + (previous.values[fieldIndex]+previous2.values[fieldIndex])/2
	case PredictorMinMotor:
		value += int32(frameDef.Sysconfig.MotorOutputLow)
	default:
		return value, errors.Errorf("Unsupported field predictor %d", predictor)
	}

	return value, nil
}

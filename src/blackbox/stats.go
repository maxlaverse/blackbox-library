package blackbox

import (
	"bytes"
	"fmt"
	"text/tabwriter"
	"time"
)

// LogStatistics represents global stats
type LogStatistics struct {
	TotalCorruptedFrames          int
	NonParsableFrames             int
	IntentionallyAbsentIterations int
	TotalFrames                   int
	Frame                         map[LogFrameType]*FrameStatistics
	Bytes                         int
	HeaderBytes                   int
	CorruptedBytes                int
	Start                         time.Time
	End                           time.Time
}

// NewLogStatistics returns an initialized LogStatistic struct
func NewLogStatistics() *LogStatistics {
	return &LogStatistics{
		Frame: map[LogFrameType]*FrameStatistics{
			LogFrameEvent: &FrameStatistics{},
			LogFrameSlow:  &FrameStatistics{},
			LogFrameIntra: &FrameStatistics{},
			LogFrameInter: &FrameStatistics{},
		},
	}
}

// FrameStatistics represents stats for one type of frame
type FrameStatistics struct {
	ValidCount   int
	Bytes        int64
	CorruptCount int
	DesyncCount  int
	SizeCount    map[int]int
}

func (s LogStatistics) String() string {
	buf := &bytes.Buffer{}

	d := s.End.Sub(s.Start)
	dTime := time.Time{}.Add(d)
	fmt.Fprintf(buf, "Log 1 of 1, start %s, end %s, duration %s\n\n", s.Start.Format("04:05.000"), s.End.Format("04:05.000"), dTime.Format("04:05.000"))
	w := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "Header bloc size:\t %d bytes\n", s.HeaderBytes)
	_, _ = fmt.Fprintf(w, "Frame bloc size:\t %d bytes\n", s.Bytes)
	_, _ = fmt.Fprintf(w, "Valid frames:\t %d (%d bytes)\n", s.TotalFrames, s.Bytes)
	_, _ = fmt.Fprintf(w, "Corrupted data:\t %d bytes\n", s.CorruptedBytes)
	_, _ = fmt.Fprintf(w, "Corrupted frames: \t %d\n", s.TotalCorruptedFrames)
	_ = w.Flush()

	buf.WriteString("\nFrame stats:\n")
	w = tabwriter.NewWriter(buf, 0, 0, 2, ' ', tabwriter.AlignRight)
	//TODO: Order those statistics
	for frameType, frameStats := range s.Frame {
		_, _ = fmt.Fprintf(
			w,
			"%s frames\t %d valid\t %d corrupt\t %d desync\t %.1f bytes avg\t %d bytes total\t %d sizes\t\n",
			string(frameType),
			frameStats.ValidCount,
			frameStats.CorruptCount,
			frameStats.DesyncCount,
			float64(frameStats.Bytes)/float64(frameStats.ValidCount),
			frameStats.Bytes,
			len(frameStats.SizeCount),
		)
	}
	_, _ = fmt.Fprintf(
		w,
		"Frames\t %d valid\t %d corrupt\t %.1f bytes avg\t %d bytes total\t\n",
		s.TotalFrames,
		s.TotalCorruptedFrames,
		float64(s.Bytes)/float64(s.TotalFrames),
		s.Bytes,
	)
	_ = w.Flush()
	fmt.Fprintf(buf, "Data rate\t %.0f Hz\t %.0f bytes/s\t\n", float64(s.TotalFrames)/d.Seconds(), float64(s.Bytes+s.HeaderBytes)/d.Seconds())

	return buf.String()
}

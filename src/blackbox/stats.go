package blackbox

import (
	"bytes"
	"fmt"
	"text/tabwriter"
)

// StatsType represents global stats
type StatsType struct {
	TotalCorruptedFrames          int
	IntentionallyAbsentIterations int
	TotalFrames                   int
	Frame                         map[LogFrameType]StatsFrameType
}

// StatsFrameType represents stats for one type of frame
type StatsFrameType struct {
	ValidCount   int
	Bytes        int64
	CorruptCount int
	SizeCount    map[int]int
	DesyncCount  int
}

func (s StatsType) String() string {
	buf := &bytes.Buffer{}

	buf.WriteString("All stats:\n")
	w := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "TotalFrames\t %d\n", s.TotalFrames)
	_, _ = fmt.Fprintf(w, "TotalCorruptedFrames\t %d\n", s.TotalCorruptedFrames)
	_ = w.Flush()

	buf.WriteString("\nFrame stats:\n")
	w = tabwriter.NewWriter(buf, 0, 0, 2, ' ', tabwriter.AlignRight)
	for frameType, frameStats := range s.Frame {
		_, _ = fmt.Fprintf(
			w,
			"%s frames\t %d valid\t %d desync\t %d bytes\t %d corrupt\t %d sizes\t\n",
			string(frameType),
			frameStats.ValidCount,
			frameStats.DesyncCount,
			frameStats.Bytes,
			frameStats.CorruptCount,
			len(frameStats.SizeCount),
		)
	}
	_ = w.Flush()

	return buf.String()
}

package blackbox

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

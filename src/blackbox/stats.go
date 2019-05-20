package blackbox

// StatsType represents global stats
type StatsType struct {
	TotalCorruptedFrames          int
	IntentionallyAbsentIterations int
	TotalFrames                   int
	Frame                         map[string]StatsFrameType
}

// StatsFrameType represents stats for one type of freame
type StatsFrameType struct {
	ValidCount   int
	Bytes        int64
	CorruptCount int
	SizeCount    map[int]int
	DesyncCount  int
}

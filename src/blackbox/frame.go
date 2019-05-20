package blackbox

// Frame represents a frame
type Frame struct {
	Type   string
	Start  int64
	End    int64
	Values []int32
}

// NewFrame returns a new frame
func NewFrame(frameType string, values []int32, start, end int64) *Frame {
	return &Frame{
		Type:   frameType,
		Values: values,
		Start:  start,
		End:    end,
	}
}

// Size returns the size in bytes of a Frame
func (f *Frame) Size() int {
	return int(f.End - f.Start)
}

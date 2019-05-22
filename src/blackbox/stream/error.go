package stream

type ReadError struct {
	reason error
}

func (e ReadError) Error() string {
	return e.reason.Error()
}

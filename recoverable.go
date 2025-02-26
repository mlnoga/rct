package rct

// Errors caused by a malformed or unexpected packet, which can be potentially be recovered by retrying the transmission
type RecoverableError struct {
	Err string
}

// Prints error to string
func (e *RecoverableError) Error() string {
	return e.Err
}

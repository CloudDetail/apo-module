package elasticsearch

type ErrInvalidIndex struct {
	Index string
}

func (e *ErrInvalidIndex) Error() string {
	return "invalid index: " + e.Index
}

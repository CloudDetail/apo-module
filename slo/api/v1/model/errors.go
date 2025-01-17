package model

import (
	"fmt"
)

type ErrInvalidSLOKey struct {
	SloKey *SLOEntryKey
}

func (e *ErrInvalidSLOKey) Error() string {
	return fmt.Sprintf("invalid slo key: %v", e.SloKey)
}

type ErrInvalidSLOType struct {
	SloType SLOType
}

func (e *ErrInvalidSLOType) Error() string {
	return fmt.Sprintf("invalid slo type: %v", e.SloType)
}

type ErrNotActiveUri struct{}

var ErrNotActiveUriError = ErrNotActiveUri{}

func (e ErrNotActiveUri) Error() string {
	return "not a active uri"
}

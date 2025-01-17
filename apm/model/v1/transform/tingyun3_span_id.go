package transform

import "fmt"

func GuidToSpanID(guid string, seq uint32) string {
	if seq == 0 {
		return guid
	}
	return fmt.Sprintf("%s-%d", guid, seq)
}

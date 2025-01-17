package transform

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"strconv"
	"unsafe"
)

func SegmentIDToSpanID(segmentID string, spanID uint32) string {
	// skywalking segmentid format:
	// 56a5e1c519ae4c76a2b8b11d92cead7f.12.16563474296430001: from TraceSegmentId
	// 56a5e1c519ae4c76a2b8b11d92cead7f: from ParentTraceSegmentId
	if len(segmentID) < 32 {
		return ""
	}
	otelSpanId := uuidTo8Bytes(swStringToUUID(segmentID, spanID))
	return hex.EncodeToString(otelSpanId[:])
}

func swStringToUUID(s string, extra uint32) (dst [16]byte) {
	// there are 2 possible formats for 's':
	// s format = 56a5e1c519ae4c76a2b8b11d92cead7f.0000000000.000000000000000000
	//            ^ start(length=32)               ^ mid(u32) ^ last(u64)
	// uid = UUID(start) XOR ([4]byte(extra) . [4]byte(uint32(mid)) . [8]byte(uint64(last)))

	// s format = 56a5e1c519ae4c76a2b8b11d92cead7f
	//            ^ start(length=32)
	// uid = UUID(start) XOR [4]byte(extra)

	if len(s) < 32 {
		return
	}

	t := unsafeGetBytes(s)
	var uid [16]byte
	_, err := hex.Decode(uid[:], t[:32])
	if err != nil {
		return uid
	}

	for i := 0; i < 4; i++ {
		uid[i] ^= byte(extra)
		extra >>= 8
	}

	if len(s) == 32 {
		return uid
	}

	index1 := bytes.IndexByte(t, '.')
	index2 := bytes.LastIndexByte(t, '.')
	if index1 != 32 || index2 < 0 {
		return
	}

	mid, err := strconv.Atoi(s[index1+1 : index2])
	if err != nil {
		return
	}

	last, err := strconv.Atoi(s[index2+1:])
	if err != nil {
		return
	}

	for i := 4; i < 8; i++ {
		uid[i] ^= byte(mid)
		mid >>= 8
	}

	for i := 8; i < 16; i++ {
		uid[i] ^= byte(last)
		last >>= 8
	}

	return uid
}

func uuidTo8Bytes(uuid [16]byte) [8]byte {
	// high bit XOR low bit
	var dst [8]byte
	for i := 0; i < 8; i++ {
		dst[i] = uuid[i] ^ uuid[i+8]
	}
	return dst
}

func unsafeGetBytes(s string) []byte {
	return (*[0x7fff0000]byte)(unsafe.Pointer(
		(*reflect.StringHeader)(unsafe.Pointer(&s)).Data),
	)[:len(s):len(s)]
}

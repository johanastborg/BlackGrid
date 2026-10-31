package bigtable

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// KeyFormat specifies how Morton codes are encoded into Bigtable row keys.
type KeyFormat int

const (
	// BinaryFormat uses 8 big-endian bytes for the Morton code.
	// Compact and perfectly preserves numeric ordering in byte order.
	BinaryFormat KeyFormat = iota

	// HexFormat uses 16 zero-padded lowercase hexadecimal characters.
	// Human-readable in Bigtable Web UI/cbt while preserving lexicographical order.
	HexFormat
)

// BuildRowKey constructs a Bigtable row key from a prefix, a Morton code, and an optional entity suffix.
func BuildRowKey(prefix string, mortonCode uint64, format KeyFormat, entityID string) []byte {
	var mortonBytes []byte
	if format == HexFormat {
		mortonBytes = []byte(fmt.Sprintf("%016x", mortonCode))
	} else {
		mortonBytes = make([]byte, 8)
		binary.BigEndian.PutUint64(mortonBytes, mortonCode)
	}

	var buf bytes.Buffer
	buf.WriteString(prefix)
	buf.Write(mortonBytes)

	if entityID != "" {
		buf.WriteString("#")
		buf.WriteString(entityID)
	}

	return buf.Bytes()
}

// BuildRangeKeys constructs the start and end row keys for a Bigtable scan interval.
// Bigtable scan ranges are typically [Start, End) or [Start, End].
// Here we generate the exact prefix bounds.
func BuildRangeKeys(prefix string, startCode, endCode uint64, format KeyFormat) (startKey, endKey []byte) {
	startKey = BuildRowKey(prefix, startCode, format, "")

	if format == HexFormat {
		// End key is formatted with endCode.
		// For an open right-bound Bigtable scan [Start, End), the successor key is often used,
		// or Bigtable's NewRange with inclusive end. Here we supply the boundary key.
		endKey = BuildRowKey(prefix, endCode, format, "\xff")
	} else {
		endMorton := make([]byte, 8)
		binary.BigEndian.PutUint64(endMorton, endCode)

		var buf bytes.Buffer
		buf.WriteString(prefix)
		buf.Write(endMorton)
		buf.WriteByte(0xFF) // inclusive boundary coverage
		endKey = buf.Bytes()
	}

	return startKey, endKey
}


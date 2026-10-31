package zcurve

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
)

// Interval represents a closed range [Start, End] of Morton codes.
type Interval struct {
	Start uint64
	End   uint64
}

// String returns a human-readable representation of the interval.
func (in Interval) String() string {
	return fmt.Sprintf("[%d, %d]", in.Start, in.End)
}

// Contains checks if a given Morton code falls within [Start, End].
func (in Interval) Contains(code uint64) bool {
	return code >= in.Start && code <= in.End
}

// Size returns the count of Morton codes within the interval.
func (in Interval) Size() uint64 {
	if in.End < in.Start {
		return 0
	}
	return (in.End - in.Start) + 1
}

// StartBytes returns the big-endian 8-byte slice suitable for Bigtable row key lexicographical ordering.
func (in Interval) StartBytes() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, in.Start)
	return buf
}

// EndBytes returns the big-endian 8-byte slice suitable for Bigtable row key lexicographical ordering.
func (in Interval) EndBytes() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, in.End)
	return buf
}

// StartHex returns the 16-character zero-padded hexadecimal string.
func (in Interval) StartHex() string {
	return hex.EncodeToString(in.StartBytes())
}

// EndHex returns the 16-character zero-padded hexadecimal string.
func (in Interval) EndHex() string {
	return hex.EncodeToString(in.EndBytes())
}

// MergeIntervals sorts intervals and merges contiguous or overlapping ones.
func MergeIntervals(intervals []Interval) []Interval {
	if len(intervals) <= 1 {
		return intervals
	}

	sort.Slice(intervals, func(i, j int) bool {
		if intervals[i].Start == intervals[j].Start {
			return intervals[i].End < intervals[j].End
		}
		return intervals[i].Start < intervals[j].Start
	})

	merged := make([]Interval, 0, len(intervals))
	curr := intervals[0]

	for i := 1; i < len(intervals); i++ {
		next := intervals[i]
		if next.Start <= curr.End || next.Start == curr.End+1 {
			if next.End > curr.End {
				curr.End = next.End
			}
		} else {
			merged = append(merged, curr)
			curr = next
		}
	}
	merged = append(merged, curr)
	return merged
}

// ConsolidateIntervals reduces intervals down to at most maxRanges by repeatedly
// merging the pair with the smallest gap between them.
func ConsolidateIntervals(intervals []Interval, maxRanges int) []Interval {
	merged := MergeIntervals(intervals)
	if maxRanges <= 0 || len(merged) <= maxRanges {
		return merged
	}

	for len(merged) > maxRanges {
		// Find pair with smallest gap (merged[i+1].Start - merged[i].End)
		minGapIdx := 0
		minGap := merged[1].Start - merged[0].End

		for i := 1; i < len(merged)-1; i++ {
			gap := merged[i+1].Start - merged[i].End
			if gap < minGap {
				minGap = gap
				minGapIdx = i
			}
		}

		// Merge minGapIdx and minGapIdx + 1
		merged[minGapIdx].End = merged[minGapIdx+1].End
		merged = append(merged[:minGapIdx+1], merged[minGapIdx+2:]...)
	}

	return merged
}


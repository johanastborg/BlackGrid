package zcurve

import (
	"bytes"
	"testing"
)

func TestMergeIntervals(t *testing.T) {
	tests := []struct {
		name     string
		input    []Interval
		expected []Interval
	}{
		{
			name:     "empty",
			input:    []Interval{},
			expected: []Interval{},
		},
		{
			name:     "single",
			input:    []Interval{{Start: 10, End: 20}},
			expected: []Interval{{Start: 10, End: 20}},
		},
		{
			name: "adjacent intervals merge",
			input: []Interval{
				{Start: 0, End: 5},
				{Start: 6, End: 10},
			},
			expected: []Interval{
				{Start: 0, End: 10},
			},
		},
		{
			name: "overlapping intervals merge",
			input: []Interval{
				{Start: 0, End: 10},
				{Start: 5, End: 15},
				{Start: 20, End: 30},
			},
			expected: []Interval{
				{Start: 0, End: 15},
				{Start: 20, End: 30},
			},
		},
		{
			name: "unsorted input",
			input: []Interval{
				{Start: 20, End: 30},
				{Start: 5, End: 10},
				{Start: 0, End: 6},
			},
			expected: []Interval{
				{Start: 0, End: 10},
				{Start: 20, End: 30},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeIntervals(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected len %d, got %d: %+v", len(tt.expected), len(got), got)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("at index %d: expected %+v, got %+v", i, tt.expected[i], got[i])
				}
			}
		})
	}
}

func TestConsolidateIntervals(t *testing.T) {
	intervals := []Interval{
		{Start: 0, End: 5},
		{Start: 10, End: 15},
		{Start: 20, End: 25},
		{Start: 40, End: 50},
	}

	consolidated := ConsolidateIntervals(intervals, 2)
	if len(consolidated) > 2 {
		t.Errorf("expected at most 2 intervals, got %d", len(consolidated))
	}

	// Verify all original numbers are still covered
	for _, in := range intervals {
		if !containsAny(consolidated, in.Start) || !containsAny(consolidated, in.End) {
			t.Errorf("consolidated lost coverage for interval %+v", in)
		}
	}
}

func TestIntervalBytesAndHex(t *testing.T) {
	in := Interval{Start: 1, End: 255}

	startBytes := in.StartBytes()
	endBytes := in.EndBytes()

	if bytes.Compare(startBytes, endBytes) >= 0 {
		t.Errorf("startBytes should be lexicographically before endBytes")
	}

	if in.StartHex() != "0000000000000001" {
		t.Errorf("unexpected start hex: %s", in.StartHex())
	}
	if in.EndHex() != "00000000000000ff" {
		t.Errorf("unexpected end hex: %s", in.EndHex())
	}
}

func containsAny(intervals []Interval, val uint64) bool {
	for _, in := range intervals {
		if in.Contains(val) {
			return true
		}
	}
	return false
}


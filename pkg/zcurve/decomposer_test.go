package zcurve

import (
	"testing"
)

func TestDecompose2D_Coverage(t *testing.T) {
	// 4 bits per dimension -> range 0..15
	bitsPerDim := 4
	d, err := NewDecomposer(2, bitsPerDim, 16)
	if err != nil {
		t.Fatalf("failed to create decomposer: %v", err)
	}

	box := Box2D{
		MinX: 2, MaxX: 5,
		MinY: 1, MaxY: 4,
	}

	intervals := d.Decompose2D(box)
	if len(intervals) == 0 {
		t.Fatal("expected at least one interval")
	}

	// Verify that EVERY point in the box is contained in at least one interval
	for x := box.MinX; x <= box.MaxX; x++ {
		for y := box.MinY; y <= box.MaxY; y++ {
			code := Encode2D(x, y)
			covered := false
			for _, in := range intervals {
				if in.Contains(code) {
					covered = true
					break
				}
			}
			if !covered {
				t.Errorf("point (%d, %d) with Morton code %d is NOT covered by any interval: %+v", x, y, code, intervals)
			}
		}
	}
}

func TestDecompose2D_WholeSpace(t *testing.T) {
	bitsPerDim := 4
	d, _ := NewDecomposer(2, bitsPerDim, 16)

	box := Box2D{
		MinX: 0, MaxX: 15,
		MinY: 0, MaxY: 15,
	}

	intervals := d.Decompose2D(box)
	if len(intervals) != 1 {
		t.Fatalf("expected 1 merged interval for whole space, got %d: %+v", len(intervals), intervals)
	}

	if intervals[0].Start != 0 || intervals[0].End != 255 {
		t.Errorf("expected interval [0, 255], got %+v", intervals[0])
	}
}

func TestDecomposeND_3DCoverage(t *testing.T) {
	bitsPerDim := 3 // range 0..7
	d, err := NewDecomposer(3, bitsPerDim, 32)
	if err != nil {
		t.Fatalf("failed to create decomposer: %v", err)
	}

	query := BoxND{
		Min: []uint64{1, 2, 0},
		Max: []uint64{3, 4, 2},
	}

	intervals := d.DecomposeND(query)
	if len(intervals) == 0 {
		t.Fatal("expected at least one interval")
	}

	for x := query.Min[0]; x <= query.Max[0]; x++ {
		for y := query.Min[1]; y <= query.Max[1]; y++ {
			for z := query.Min[2]; z <= query.Max[2]; z++ {
				code := Encode3D(uint32(x), uint32(y), uint32(z))
				covered := false
				for _, in := range intervals {
					if in.Contains(code) {
						covered = true
						break
					}
				}
				if !covered {
					t.Errorf("3D point (%d, %d, %d) code %d not covered", x, y, z, code)
				}
			}
		}
	}
}

func TestDecompose2D_MaxRangesBudget(t *testing.T) {
	bitsPerDim := 8
	maxRanges := 8
	d, _ := NewDecomposer(2, bitsPerDim, maxRanges)

	// A diagonal-like or sparse box that generates many disjoint intervals
	box := Box2D{
		MinX: 10, MaxX: 100,
		MinY: 15, MaxY: 120,
	}

	intervals := d.Decompose2D(box)
	if len(intervals) > maxRanges {
		t.Errorf("expected at most %d intervals, got %d", maxRanges, len(intervals))
	}
}


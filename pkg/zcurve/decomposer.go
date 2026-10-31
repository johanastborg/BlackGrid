package zcurve

import (
	"fmt"
)

// Box2D represents an axis-aligned 2D bounding box with integer coordinates.
type Box2D struct {
	MinX, MaxX uint32
	MinY, MaxY uint32
}

// BoxND represents an axis-aligned N-dimensional bounding box.
type BoxND struct {
	Min []uint64
	Max []uint64
}

// Validate checks that min <= max for all dimensions.
func (b BoxND) Validate() error {
	if len(b.Min) != len(b.Max) {
		return fmt.Errorf("dimension mismatch: min has %d, max has %d", len(b.Min), len(b.Max))
	}
	if len(b.Min) == 0 {
		return ErrInvalidDimensions
	}
	for i := range b.Min {
		if b.Min[i] > b.Max[i] {
			return fmt.Errorf("dimension %d: min (%d) > max (%d)", i, b.Min[i], b.Max[i])
		}
	}
	return nil
}

// Decomposer configures the Morton space decomposition parameters.
type Decomposer struct {
	Dimensions int
	BitsPerDim int
	MaxRanges  int
	MaxDepth   int // Maximum tree depth to search before including the cell
}

// NewDecomposer creates a Decomposer with sensible defaults.
func NewDecomposer(dimensions, bitsPerDim, maxRanges int) (*Decomposer, error) {
	if dimensions < 1 {
		return nil, ErrInvalidDimensions
	}
	if dimensions*bitsPerDim > 64 {
		return nil, fmt.Errorf("%w: %d * %d = %d", ErrTooManyBits, dimensions, bitsPerDim, dimensions*bitsPerDim)
	}
	if maxRanges <= 0 {
		maxRanges = 64
	}

	maxDepth := bitsPerDim
	// Bound maximum search tree depth based on dimensionality to prevent combinatorial explosion:
	// Total branching exponent D * depth <= 24
	maxTreeBits := 24
	if dimensions > 1 {
		suggestedDepth := maxTreeBits / dimensions
		if suggestedDepth < maxDepth {
			maxDepth = suggestedDepth
		}
	}
	if maxDepth < 4 {
		maxDepth = 4
	}

	return &Decomposer{
		Dimensions: dimensions,
		BitsPerDim: bitsPerDim,
		MaxRanges:  maxRanges,
		MaxDepth:   maxDepth,
	}, nil
}

// Decompose2D decomposes a 2D bounding box into a set of 1D Morton intervals.
func (d *Decomposer) Decompose2D(box Box2D) []Interval {
	boxND := BoxND{
		Min: []uint64{uint64(box.MinX), uint64(box.MinY)},
		Max: []uint64{uint64(box.MaxX), uint64(box.MaxY)},
	}
	return d.DecomposeND(boxND)
}

// DecomposeND decomposes an N-dimensional bounding box into a set of 1D Morton intervals.
func (d *Decomposer) DecomposeND(query BoxND) []Interval {
	if err := query.Validate(); err != nil || len(query.Min) != d.Dimensions {
		return nil
	}

	// Clamp query to max representable coordinate
	var maxCoord uint64
	if d.BitsPerDim == 64 {
		maxCoord = ^uint64(0)
	} else {
		maxCoord = (uint64(1) << d.BitsPerDim) - 1
	}

	clampedQuery := BoxND{
		Min: make([]uint64, d.Dimensions),
		Max: make([]uint64, d.Dimensions),
	}
	for i := 0; i < d.Dimensions; i++ {
		clampedQuery.Min[i] = query.Min[i]
		if clampedQuery.Min[i] > maxCoord {
			clampedQuery.Min[i] = maxCoord
		}
		clampedQuery.Max[i] = query.Max[i]
		if clampedQuery.Max[i] > maxCoord {
			clampedQuery.Max[i] = maxCoord
		}
		if clampedQuery.Min[i] > clampedQuery.Max[i] {
			return nil
		}
	}

	var intervals []Interval
	d.decomposeCell(0, 0, make([]uint64, d.Dimensions), &clampedQuery, &intervals)

	intervals = MergeIntervals(intervals)
	if d.MaxRanges > 0 && len(intervals) > d.MaxRanges {
		intervals = ConsolidateIntervals(intervals, d.MaxRanges)
	}

	return intervals
}

func (d *Decomposer) decomposeCell(
	depth int,
	prefixCode uint64,
	cellCoords []uint64,
	query *BoxND,
	results *[]Interval,
) {
	dim := d.Dimensions
	remainingBits := d.BitsPerDim - depth

	// Calculate spatial bounding box for the current cell
	cellMin := make([]uint64, dim)
	cellMax := make([]uint64, dim)
	span := uint64(1) << remainingBits
	for i := 0; i < dim; i++ {
		cellMin[i] = cellCoords[i] << remainingBits
		cellMax[i] = cellMin[i] + span - 1
	}

	// 1. Check if cell is disjoint with query
	for i := 0; i < dim; i++ {
		if cellMax[i] < query.Min[i] || cellMin[i] > query.Max[i] {
			return // Disjoint, prune
		}
	}

	// 2. Check if cell is completely contained within query
	contained := true
	for i := 0; i < dim; i++ {
		if cellMin[i] < query.Min[i] || cellMax[i] > query.Max[i] {
			contained = false
			break
		}
	}

	shift := remainingBits * dim
	startCode := prefixCode << shift
	endCode := startCode | ((uint64(1) << shift) - 1)

	// Early termination: contained, reached max search depth, or generated enough candidate intervals
	if contained || depth >= d.MaxDepth || (d.MaxRanges > 0 && len(*results) >= d.MaxRanges*4) {
		*results = append(*results, Interval{Start: startCode, End: endCode})
		return
	}

	// 3. Partially overlapping: subdivide into 2^D child cells
	numChildren := 1 << dim
	nextCellCoords := make([]uint64, dim)

	for childIdx := 0; childIdx < numChildren; childIdx++ {
		childPrefix := (prefixCode << dim) | uint64(childIdx)
		for i := 0; i < dim; i++ {
			bit := (uint64(childIdx) >> i) & 1
			nextCellCoords[i] = (cellCoords[i] << 1) | bit
		}
		d.decomposeCell(depth+1, childPrefix, nextCellCoords, query, results)
	}
}

package zcurve

import (
	"errors"
	"fmt"
)

var (
	// ErrTooManyBits indicates total bits (dim * bitsPerDim) exceeds 64.
	ErrTooManyBits = errors.New("total bits (dimensions * bitsPerDim) exceeds 64 bits")
	// ErrInvalidDimensions indicates dimension count < 1.
	ErrInvalidDimensions = errors.New("dimensions must be at least 1")
	// ErrCoordinateOutOfBounds indicates a coordinate exceeds the bit length capacity.
	ErrCoordinateOutOfBounds = errors.New("coordinate value exceeds bitsPerDim limit")
)

// EncodeND interleaves an arbitrary number of dimensions into a 64-bit Morton code.
// dims * bitsPerDim must not exceed 64.
func EncodeND(coords []uint64, bitsPerDim int) (uint64, error) {
	d := len(coords)
	if d < 1 {
		return 0, ErrInvalidDimensions
	}
	if d*bitsPerDim > 64 {
		return 0, fmt.Errorf("%w: %d * %d = %d", ErrTooManyBits, d, bitsPerDim, d*bitsPerDim)
	}

	var maxVal uint64
	if bitsPerDim == 64 {
		maxVal = ^uint64(0)
	} else {
		maxVal = (uint64(1) << bitsPerDim) - 1
	}

	for i, c := range coords {
		if c > maxVal {
			return 0, fmt.Errorf("%w: coord[%d]=%d exceeds max %d", ErrCoordinateOutOfBounds, i, c, maxVal)
		}
	}

	// Fast paths for 2D and 3D
	if d == 2 && bitsPerDim <= 32 {
		return Encode2D(uint32(coords[0]), uint32(coords[1])), nil
	}
	if d == 3 && bitsPerDim <= 21 {
		return Encode3D(uint32(coords[0]), uint32(coords[1]), uint32(coords[2])), nil
	}

	var code uint64
	for b := 0; b < bitsPerDim; b++ {
		for i := 0; i < d; i++ {
			bit := (coords[i] >> b) & 1
			code |= bit << (b*d + i)
		}
	}
	return code, nil
}

// DecodeND unpacks a 64-bit Morton code into d coordinates of bitsPerDim each.
func DecodeND(code uint64, d int, bitsPerDim int) ([]uint64, error) {
	if d < 1 {
		return nil, ErrInvalidDimensions
	}
	if d*bitsPerDim > 64 {
		return nil, fmt.Errorf("%w: %d * %d = %d", ErrTooManyBits, d, bitsPerDim, d*bitsPerDim)
	}

	if d == 2 && bitsPerDim <= 32 {
		x, y := Decode2D(code)
		return []uint64{uint64(x), uint64(y)}, nil
	}
	if d == 3 && bitsPerDim <= 21 {
		x, y, z := Decode3D(code)
		return []uint64{uint64(x), uint64(y), uint64(z)}, nil
	}

	coords := make([]uint64, d)
	for b := 0; b < bitsPerDim; b++ {
		for i := 0; i < d; i++ {
			bit := (code >> (b*d + i)) & 1
			coords[i] |= bit << b
		}
	}
	return coords, nil
}

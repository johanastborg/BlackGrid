package zcurve

import (
	"math/rand"
	"testing"
)

func TestEncodeDecodeND(t *testing.T) {
	dimensionsList := []int{1, 2, 3, 4, 8}
	rng := rand.New(rand.NewSource(123))

	for _, dim := range dimensionsList {
		bitsPerDim := 64 / dim
		maxCoord := (uint64(1) << bitsPerDim) - 1

		for i := 0; i < 500; i++ {
			coords := make([]uint64, dim)
			for d := 0; d < dim; d++ {
				coords[d] = rng.Uint64() & maxCoord
			}

			code, err := EncodeND(coords, bitsPerDim)
			if err != nil {
				t.Fatalf("dim %d: unexpected EncodeND error: %v", dim, err)
			}

			recovered, err := DecodeND(code, dim, bitsPerDim)
			if err != nil {
				t.Fatalf("dim %d: unexpected DecodeND error: %v", dim, err)
			}

			for d := 0; d < dim; d++ {
				if recovered[d] != coords[d] {
					t.Fatalf("dim %d, iter %d: dim %d coord mismatch: got %d, want %d",
						dim, i, d, recovered[d], coords[d])
				}
			}
		}
	}
}

func TestEncodeND_Validation(t *testing.T) {
	// Zero dimensions
	if _, err := EncodeND([]uint64{}, 10); err == nil {
		t.Error("expected error for 0 dimensions")
	}

	// Bit overflow (> 64)
	if _, err := EncodeND([]uint64{1, 2, 3}, 25); err == nil {
		t.Error("expected error for 3 * 25 = 75 bits")
	}

	// Out of bounds coordinate
	if _, err := EncodeND([]uint64{16, 2}, 4); err == nil { // 4 bits max is 15
		t.Error("expected error when coordinate exceeds max 15")
	}
}


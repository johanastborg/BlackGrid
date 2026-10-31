package zcurve

import (
	"math/rand"
	"testing"
)

func TestEncodeDecode3D_KnownValues(t *testing.T) {
	tests := []struct {
		x, y, z uint32
		want    uint64
	}{
		{0, 0, 0, 0},
		{1, 0, 0, 1},
		{0, 1, 0, 2},
		{0, 0, 1, 4},
		{1, 1, 1, 7},
		{MaxCoordinate3D, MaxCoordinate3D, MaxCoordinate3D, (uint64(1) << 63) - 1},
	}

	for _, tt := range tests {
		got := Encode3D(tt.x, tt.y, tt.z)
		if got != tt.want {
			t.Errorf("Encode3D(%d, %d, %d) = %d (0x%x); want %d (0x%x)", tt.x, tt.y, tt.z, got, got, tt.want, tt.want)
		}
		rx, ry, rz := Decode3D(got)
		if rx != tt.x || ry != tt.y || rz != tt.z {
			t.Errorf("Decode3D(%d) = (%d, %d, %d); want (%d, %d, %d)", got, rx, ry, rz, tt.x, tt.y, tt.z)
		}
	}
}

func TestEncodeDecode3D_RoundTripRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	for i := 0; i < 10000; i++ {
		x := rng.Uint32() & MaxCoordinate3D
		y := rng.Uint32() & MaxCoordinate3D
		z := rng.Uint32() & MaxCoordinate3D

		code := Encode3D(x, y, z)
		rx, ry, rz := Decode3D(code)

		if rx != x || ry != y || rz != z {
			t.Fatalf("Iteration %d: mismatch for (%d, %d, %d) -> code %d -> got (%d, %d, %d)", i, x, y, z, code, rx, ry, rz)
		}
	}
}


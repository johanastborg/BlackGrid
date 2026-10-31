package zcurve

import (
	"math/rand"
	"testing"
)

func TestEncodeDecode2D_KnownValues(t *testing.T) {
	tests := []struct {
		x, y uint32
		want uint64
	}{
		{0, 0, 0},
		{1, 0, 1},
		{0, 1, 2},
		{1, 1, 3},
		{2, 0, 4},
		{0, 2, 8},
		{2, 2, 12}, // 0b1100
		{3, 3, 15}, // 0b1111
		{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
	}

	for _, tt := range tests {
		got := Encode2D(tt.x, tt.y)
		if got != tt.want {
			t.Errorf("Encode2D(%d, %d) = %d (0x%x); want %d (0x%x)", tt.x, tt.y, got, got, tt.want, tt.want)
		}
		rx, ry := Decode2D(got)
		if rx != tt.x || ry != tt.y {
			t.Errorf("Decode2D(%d) = (%d, %d); want (%d, %d)", got, rx, ry, tt.x, tt.y)
		}
	}
}

func TestEncodeDecode2D_RoundTripRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 10000; i++ {
		x := rng.Uint32()
		y := rng.Uint32()

		code := Encode2D(x, y)
		rx, ry := Decode2D(code)

		if rx != x || ry != y {
			t.Fatalf("Iteration %d: mismatch for (%d, %d) -> code %d -> got (%d, %d)", i, x, y, code, rx, ry)
		}
	}
}

func BenchmarkEncode2D(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode2D(uint32(i), uint32(i*2))
	}
}

func BenchmarkDecode2D(b *testing.B) {
	code := Encode2D(12345, 67890)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode2D(code)
	}
}


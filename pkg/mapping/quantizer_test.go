package mapping

import (
	"math"
	"testing"
)

func TestQuantizeDequantize(t *testing.T) {
	dim := &Dimension{
		Name:   "latitude",
		Type:   TypeFloat64,
		MinVal: -90.0,
		MaxVal: 90.0,
		Bits:   16, // 0..65535
	}

	tests := []float64{-90.0, -45.0, 0.0, 45.0, 90.0}

	for _, v := range tests {
		q := Quantize(dim, v)
		deq := Dequantize(dim, q)

		// Max error is range / (2^bits - 1)
		maxErr := (dim.MaxVal - dim.MinVal) / 65535.0
		diff := math.Abs(deq - v)
		if diff > maxErr+1e-6 {
			t.Errorf("Quantize/Dequantize(%v): got %v, err %v > maxErr %v", v, deq, diff, maxErr)
		}
	}
}

func TestQuantizeClamping(t *testing.T) {
	dim := &Dimension{
		Name:   "lon",
		Type:   TypeFloat64,
		MinVal: -180.0,
		MaxVal: 180.0,
		Bits:   8, // max 255
	}

	if q := Quantize(dim, -200.0); q != 0 {
		t.Errorf("expected underflow clamping to 0, got %d", q)
	}

	if q := Quantize(dim, 250.0); q != 255 {
		t.Errorf("expected overflow clamping to 255, got %d", q)
	}
}


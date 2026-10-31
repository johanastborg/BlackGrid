package mapping

import (
	"fmt"
	"math"
)

// Quantizer maps continuous or discrete values into integer coordinates for Morton encoding.
type Quantizer struct{}

// Quantize maps a value in [dim.MinVal, dim.MaxVal] to an unsigned integer [0, 2^dim.Bits - 1].
func Quantize(dim *Dimension, val float64) uint64 {
	if val <= dim.MinVal {
		return 0
	}
	if val >= dim.MaxVal {
		if dim.Bits == 64 {
			return ^uint64(0)
		}
		return (uint64(1) << dim.Bits) - 1
	}

	maxDiscrete := float64((uint64(1) << dim.Bits) - 1)
	ratio := (val - dim.MinVal) / (dim.MaxVal - dim.MinVal)
	quantized := math.Round(ratio * maxDiscrete)

	if quantized < 0 {
		return 0
	}
	if quantized > maxDiscrete {
		return uint64(maxDiscrete)
	}
	return uint64(quantized)
}

// Dequantize maps an unsigned integer back to [dim.MinVal, dim.MaxVal].
func Dequantize(dim *Dimension, quantized uint64) float64 {
	maxDiscrete := float64((uint64(1) << dim.Bits) - 1)
	if float64(quantized) >= maxDiscrete {
		return dim.MaxVal
	}
	ratio := float64(quantized) / maxDiscrete
	return dim.MinVal + ratio*(dim.MaxVal-dim.MinVal)
}

// QuantizeRange converts a continuous query range [minVal, maxVal] to integer coordinate bounds.
func QuantizeRange(dim *Dimension, minVal, maxVal float64) (uint64, uint64, error) {
	if minVal > maxVal {
		return 0, 0, fmt.Errorf("invalid range for dimension %q: min %v > max %v", dim.Name, minVal, maxVal)
	}

	qMin := Quantize(dim, minVal)
	qMax := Quantize(dim, maxVal)

	return qMin, qMax, nil
}


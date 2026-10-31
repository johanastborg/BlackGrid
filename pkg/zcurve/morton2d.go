package zcurve

// Encode2D interleaves the bits of two 32-bit unsigned integers into a 64-bit Morton code.
// Bits of x are at even indices (0, 2, 4, ...), bits of y are at odd indices (1, 3, 5, ...).
func Encode2D(x, y uint32) uint64 {
	return spread2D(x) | (spread2D(y) << 1)
}

// Decode2D splits a 64-bit Morton code into two 32-bit unsigned integers (x, y).
func Decode2D(code uint64) (x, y uint32) {
	x = compact2D(code)
	y = compact2D(code >> 1)
	return x, y
}

func spread2D(x uint32) uint64 {
	v := uint64(x)
	v = (v | (v << 16)) & 0x0000FFFF0000FFFF
	v = (v | (v << 8)) & 0x00FF00FF00FF00FF
	v = (v | (v << 4)) & 0x0F0F0F0F0F0F0F0F
	v = (v | (v << 2)) & 0x3333333333333333
	v = (v | (v << 1)) & 0x5555555555555555
	return v
}

func compact2D(v uint64) uint32 {
	v &= 0x5555555555555555
	v = (v | (v >> 1)) & 0x3333333333333333
	v = (v | (v >> 2)) & 0x0F0F0F0F0F0F0F0F
	v = (v | (v >> 4)) & 0x00FF00FF00FF00FF
	v = (v | (v >> 8)) & 0x0000FFFF0000FFFF
	v = (v | (v >> 16)) & 0x00000000FFFFFFFF
	return uint32(v)
}


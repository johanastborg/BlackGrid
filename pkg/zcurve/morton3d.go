package zcurve

const (
	// MaxCoordinate3D is the maximum 21-bit coordinate (2^21 - 1 = 2,097,151)
	MaxCoordinate3D uint32 = (1 << 21) - 1
)

// Encode3D interleaves the bits of three 21-bit unsigned integers into a 64-bit Morton code.
// Bits of x are at positions 0, 3, 6, ...
// Bits of y are at positions 1, 4, 7, ...
// Bits of z are at positions 2, 5, 8, ...
func Encode3D(x, y, z uint32) uint64 {
	return spread3D(x) | (spread3D(y) << 1) | (spread3D(z) << 2)
}

// Decode3D splits a 64-bit Morton code into three 21-bit unsigned integers (x, y, z).
func Decode3D(code uint64) (x, y, z uint32) {
	x = compact3D(code)
	y = compact3D(code >> 1)
	z = compact3D(code >> 2)
	return x, y, z
}

func spread3D(x uint32) uint64 {
	v := uint64(x) & 0x1FFFFF
	v = (v | (v << 32)) & 0x001F00000000FFFF
	v = (v | (v << 16)) & 0x001F0000FF0000FF
	v = (v | (v << 8)) & 0x100F00F00F00F00F
	v = (v | (v << 4)) & 0x10C30C30C30C30C3
	v = (v | (v << 2)) & 0x1249249249249249
	return v
}

func compact3D(v uint64) uint32 {
	v &= 0x1249249249249249
	v = (v | (v >> 2)) & 0x10C30C30C30C30C3
	v = (v | (v >> 4)) & 0x100F00F00F00F00F
	v = (v | (v >> 8)) & 0x001F0000FF0000FF
	v = (v | (v >> 16)) & 0x001F00000000FFFF
	v = (v | (v >> 32)) & 0x00000000001FFFFF
	return uint32(v)
}


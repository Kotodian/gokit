package utils

var (
	_ = uint64(1)
)

func SetBit(i uint, pos uint) uint {
	return i | (0x1 << pos)
}

func ClearBit(i uint, pos uint) uint {
	return i & (^(0x1 << (pos)))
}

func GetBit(i uint, pos uint) uint {
	return (i >> pos) & 0x1
}

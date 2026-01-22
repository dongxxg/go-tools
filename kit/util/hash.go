package util

import (
	"hash/crc32"
	"io"
)

// GenHash 计算 一致性hash
func GenHash(id string) uint32 {
	ieee := crc32.NewIEEE()
	_, _ = io.WriteString(ieee, id)
	return ieee.Sum32()
}
func GetHashNo(id string, total int) int {
	n := GenHash(id)
	return int(n) % total
}

package util

import (
	"golang.org/x/crypto/sha3"
)

func Sha3(data []byte) []byte {
	// 使用SHA3-256对数据进行哈希
	hash := sha3.New256()
	hash.Write(data)
	hashSum := hash.Sum(nil)
	return hashSum
}

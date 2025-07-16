package util

import (
	"bytes"
)

func Join(pBytes ...[]byte) []byte {
	return JoinStr(pBytes, []byte(""))
}

func JoinStr(pBytes [][]byte, spe []byte) []byte {
	return bytes.Join(pBytes, spe)
}

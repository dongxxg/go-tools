package crypto

import (
	cryptorand "crypto/rand"
	"math/big"
)

const (
	numStr  = "0123456789"
	charStr = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	specStr = "+=-@#~,.[]()!%^*$"
)

const (
	passwdNum     = "num"
	passwdChar    = "char"
	passwdMix     = "mix"
	passwdAdvance = "advance"
)

// GenPasswdNum generates a numeric-only password of the given length.
func GenPasswdNum(l int) string {
	return generatePasswd(l, passwdNum)
}

// GenPasswdChar generates an alphabetic-only password of the given length.
func GenPasswdChar(l int) string {
	return generatePasswd(l, passwdChar)
}

// GenPasswdMix generates an alphanumeric password of the given length.
func GenPasswdMix(l int) string {
	return generatePasswd(l, passwdMix)
}

// GenPasswdAdvance generates an alphanumeric password with special characters.
func GenPasswdAdvance(l int) string {
	return generatePasswd(l, passwdAdvance)
}

func generatePasswd(length int, charset string) string {
	if length <= 0 {
		length = 8
	}

	var sourceStr string
	switch charset {
	case passwdNum:
		sourceStr = numStr
	case passwdChar:
		sourceStr = charStr
	case passwdMix:
		sourceStr = numStr + charStr
	case passwdAdvance:
		sourceStr = numStr + charStr + specStr
	default:
		sourceStr = numStr
	}

	sourceLen := big.NewInt(int64(len(sourceStr)))
	passwd := make([]byte, length)
	for i := 0; i < length; i++ {
		idx, err := cryptorand.Int(cryptorand.Reader, sourceLen)
		if err != nil {
			// fall back to the first character on the very unlikely RNG failure
			passwd[i] = sourceStr[0]
			continue
		}
		passwd[i] = sourceStr[idx.Int64()]
	}
	return string(passwd)
}

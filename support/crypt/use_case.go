package crypt

// Aes存储秘钥

// crypto 加密算法
func StorageEncrypt(srcStr []byte, crypto int, signKey []byte) ([]byte, error) {
	switch crypto {
	case AesCbc:
		return cbcEncrypt(srcStr, signKey, iV)
	default:
		return srcStr, nil
	}
}

func StorageDecrypt(encrypt []byte, crypto int, signKey []byte) ([]byte, error) {
	switch crypto {
	case AesCbc:
		return cbcDecrypt(encrypt, signKey, iV)
	default:
		return encrypt, nil
	}
}

func CommonEncrypt(srcStr []byte, crypto int, key []byte) ([]byte, error) {
	switch crypto {
	case AesCbc:
		return cbcEncrypt(srcStr, key, iV)
	default:
		return srcStr, nil
	}
}

func CommonDecrypt(encryptStr []byte, crypto int, key []byte) ([]byte, error) {
	switch crypto {
	case AesCbc:
		return cbcDecrypt(encryptStr, key, iV)
	default:
		return encryptStr, nil
	}
}

package common

// 从lib复制过来
import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

// 入参都为string, 出参string, error
func AesDecryptString(crypted, key string) (string, error) {
	cryptedbyte, err := base64.StdEncoding.DecodeString(crypted)
	if err != nil {
		return "", err
	}
	newKey := keyFormat(key)
	decryptbyte, err := AesDecrypt(cryptedbyte, []byte(newKey))
	if err != nil {
		return "", err
	}
	return string(decryptbyte), nil
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = ZeroUnPadding(origData)
	return origData, nil
}

// 说明：AES数据块长度为128位，所以IV长度需要为16个字符（ECB模式不用IV），
// 密钥根据指定密钥位数分别为16、24、32个字符，IV与密钥超过长度则截取，不足则在末尾填充'\0'补足
func keyFormat(key string) string {
	newKey := key
	for i := len(key); i < 16; i++ {
		newKey += string([]byte{0})
	}
	return newKey[:16]
}

func ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return bytes.TrimRight(origData[:(length-unpadding)], string(byte(0)))
}

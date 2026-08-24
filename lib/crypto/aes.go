package crypto

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

// 入参都为string, 出参string, error
func AesEncryptString(origstr, key string) (string, error) {
	newKey := keyFormat(key)
	cryptedbyte, err := AesEncrypt([]byte(origstr), []byte(newKey))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(cryptedbyte), nil
}

func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	//origData = PKCS5Padding(origData, blockSize)
	origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	// origData = PKCS5UnPadding(origData)
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

func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	// return origData[:(length - unpadding)]
	return bytes.TrimRight(origData[:(length-unpadding)], string(byte(0)))
}

// 其实是pkcs7
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

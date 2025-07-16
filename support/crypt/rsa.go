/**
 * @author  zhaoliang.liang
 * @date  2024/2/29 0029 13:44
 */

package crypt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// RSAEncrypt 使用 RSA 公钥加密数据
func RSAEncrypt(pubKey *rsa.PublicKey, data []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
}

// RSADecrypt 使用 RSA 私钥解密数据
func RSADecrypt(privKey *rsa.PrivateKey, cipherText []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, privKey, cipherText)
}

func SignData(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func VerifySignature(data []byte, signature []byte, publicKey *rsa.PublicKey) error {
	hashed := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
}
func LoadPrivateKeyFromFile(privKeyPath string) (*rsa.PrivateKey, error) {
	privKeyBytes, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key := privKey.(*rsa.PrivateKey)

	return key, nil
}

func LoadPrivateKeyFromStr(privKeyStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privKeyStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key := privKey.(*rsa.PrivateKey)

	return key, nil
}

func LoadPublicKeyFromFile(pubKeyPath string) (*rsa.PublicKey, error) {
	pubKeyBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pubKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPubKey, nil
}

func LoadPublicKeyFromStr(pubKeyStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubKeyStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPubKey, nil
}

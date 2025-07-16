/**
 * @author  zhaoliang.liang
 * @date  2024/2/29 0029 15:08
 */

package crypt

import "encoding/base64"

// EncodeBase64 将字节数组转换为Base64字符串
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 将Base64字符串解码为字节数组
func DecodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

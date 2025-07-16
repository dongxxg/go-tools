package util

import "strings"

const OSSDOMAIN_ACCEL = "file.plaso.cn"
const OSSDOMAIN_CDN = "filecdn.plaso.cn"

// 获取文件相对于课堂的路径，即文件名
func GetOssRelativePath(path string) string {
	if strings.Contains(path, ".aliyuncs.com") || strings.Contains(path, OSSDOMAIN_ACCEL) || strings.Contains(path, OSSDOMAIN_CDN) {
		infoArr := strings.Split(path, "/")
		fileName := infoArr[len(infoArr)-1]
		return fileName
	} else {
		return ""
	}
}

func GetRelativePath(value interface{}) interface{} {
	if path, ok := value.(string); ok {
		if strings.Contains(path, "http") {
			v := GetOssRelativePath(path)
			if len(v) != 0 {
				return v
			}
		}
	}
	return value
}

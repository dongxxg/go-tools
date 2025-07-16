package util

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"support/logger"

	"github.com/pkg/errors"
)

const (
	gzipID1     = 0x1f
	gzipID2     = 0x8b
	gzipDeflate = 8
)

const DEFAULT_PERM = 0755

// JoinPath
//
//	@Description: 拼接文件路径，对路径进行优化处理，保证路径的正确性
//
// 防止直接把path拼接出现连续双斜杆或者没有斜杠的问题
//
//	@param path1:前一个路径
//	@param path2：后一个路径
//	@return string :最终路径
func JoinPath(path1 string, path2 string) string {
	if !strings.HasSuffix(path1, "/") {
		path1 = path1 + "/"
	}
	if strings.HasPrefix(path2, "/") {
		path2 = strings.TrimPrefix(path2, "/")
	}
	return path1 + path2
}

// 针对/abc/d/efg/h.xx ，解析h.xx之前的部分,带/
func PrePath(pathstr string) string {
	basename := filepath.Base(pathstr)
	length := len(pathstr) - len(basename)
	return pathstr[:length]
}

func IsExisted(localFilePath string) (bool, error) {
	_, err := os.Stat(localFilePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetContent(localPath string) ([]byte, error) {
	b, err := ioutil.ReadFile(localPath)
	if err != nil {
		logger.Warn("Error can not get file %v", err)
		return nil, err
	}
	return b, nil
}

func Ungzip(gzipFile string) ([]byte, error) {
	// 打开gzip文件
	fr, err := os.Open(gzipFile)
	if err != nil {
		log.Println(err)
	}
	defer fr.Close()
	// 创建gzip.Reader
	gr, err := gzip.NewReader(fr)
	if err == gzip.ErrHeader {
		//不是gzip文件
		return GetContent(gzipFile)
	} else if err != nil {
		return nil, err
	}
	defer gr.Close()
	res, err2 := io.ReadAll(gr)
	return res, err2
}

func IsGzip(localPath string) (bool, error) {
	data, err := ioutil.ReadFile(localPath)
	if err != nil {
		return false, err
	}
	if data[0] == gzipID1 && data[1] == gzipID2 && data[2] == gzipDeflate {
		//检验文件是否为gzip文件，可参考http://www.onicos.com/staff/iz/formats/gzip.html或直接查看gunzip源码
		return true, nil
	} else {
		return false, nil
	}
}

func WriteBytesFile(localPath string, m []byte) error {
	if err := PreCheckDir(localPath); err != nil {
		return err
	}
	err2 := ioutil.WriteFile(localPath, m, DEFAULT_PERM)
	if err2 != nil {
		logger.Error(fmt.Sprintf("write file to %s failed. err = %s", localPath, err2.Error()))
		return err2
	} else {
		return nil
	}
}

/**
* 将对象转为json后下载到本地
* @param gzipFlag 表示是否压缩后上传
**/
func WriteJsonFile(object interface{}, localPath string, gzipFlag bool) error {
	if err := PreCheckDir(localPath); err != nil {
		return err
	}
	data, errs := json.Marshal(object)
	if errs != nil {
		logger.Error("write json file %s failed. err=%s", localPath, errs)
		return errs
	}
	res := data
	if gzipFlag {
		var b bytes.Buffer
		w := gzip.NewWriter(&b)
		defer w.Close()
		w.Write(data)
		w.Close()
		res = b.Bytes()
	}
	// basename := filepath.Base(localPath)
	err := os.WriteFile(localPath, res, DEFAULT_PERM)
	if err != nil {
		logger.Error("write json to file %s failed. err=%s", localPath, err)
		return err
	}
	return nil
}

/**
* 检测localPath所在的目录是否存在，如果不存在就创建目录
**/
func PreCheckDir(localPath string) error {
	fileDir := filepath.Dir(localPath)
	info, err := os.Stat(fileDir)
	if err != nil || !info.IsDir() {
		err1 := os.MkdirAll(fileDir, DEFAULT_PERM)
		if err1 != nil {
			logger.Error(fmt.Sprintf("mkdir filedir %s failed. err = %s", localPath, err1.Error()))
			return err1
		}
	}
	return nil
}

func ToRelativePath(fullPath string) string {
	if strings.Contains(fullPath, ".aliyuncs.com") || strings.Contains(fullPath, "file.plaso.cn") {
		return filepath.Base(fullPath)
	} else {
		return fullPath
	}
}

// CopyFile 本地文件复制
// @src 源文件地址
// @dest 目标文件地址；不会自动创建路径上的目录（返回错误：The system cannot find the path specified）
// @perm 目标文件的权限
// @ret 返回错误信息
func CopyFile(src, dest string, perm fs.FileMode) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return errors.Wrap(err, "read src file")
	}
	err = ioutil.WriteFile(dest, input, perm)
	if err != nil {
		return errors.Wrap(err, "write dest file")
	}
	return nil
}

// GetFileNameWithoutExt
//
//	@Description: 获取文件名（去除了后缀）
//
// @param filename
// @return string
// e.g. 入参 a/aHAHA哈哈哈a134.jpg  返回 aHAHA哈哈哈a134
func GetFileNameWithoutExt(filename string) string {
	if len(filename) == 0 {
		return ""
	}
	basename := filepath.Base(filename)
	ext := filepath.Ext(filename)
	res := basename[0 : len(basename)-len(ext)]
	return res
}

func GetContentDisposition(fileName string) string {
	return fmt.Sprintf(
		"attachment; filename=%s;filename*=UTF-8''%s",
		url.PathEscape(fileName),
		url.PathEscape(fileName),
	)
}

func CheckFileNameHasSpecialChar(fileName string) bool {
	if strings.ContainsAny(fileName, "/\\:*?\"<>|") {
		return true
	}
	return false
}

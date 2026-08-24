package filex

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"

	"unitechs.com/unios-dice/uni-base/core/log"
)

const bufSize = 1024

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// SearchFile Search a file in paths.
func SearchFile(filename string, paths ...string) (fullpath string, err error) {
	isExist := false
	for _, path := range paths {
		fullpath = filepath.Join(path, filename)
		isExist, err = PathExists(fullpath)
		if err == nil && isExist {
			return
		}
	}

	fullpath = ""
	err = errors.New(filename + " not found in paths")
	return
}

func MakeDir(dir string) error {
	exist, err := PathExists(dir)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	err = os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// IsFile checks whether the path is a file,
// it returns false when it's a directory or does not exist.
func IsFile(fp string) bool {
	f, e := os.Stat(fp)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

// Unlink deletes a file
func Unlink(fp string) error {
	return os.Remove(fp)
}

// Rename 这个系统操作只能用于操作系统同一分区之间的文件rename
func Rename(src, target string) error {
	return os.Rename(src, target)
}

func WriteToFileByte(file string, b []byte) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("open file %s failed %s", file, err.Error())
		return err
	}
	defer f.Close()
	_, err = f.Write(b)
	if err != nil {
		log.Errorf("write file %s failed %s", file, err.Error())
		return err
	}
	return nil
}

func AppendToFileByte(file string, b []byte) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// GrepFile like command grep -E
// for example: GrepFile(`^hello`, "hello.txt")
// \n is striped while read
func GrepFile(patten, filename string) (lines []string, err error) {
	re, err := regexp.Compile(patten)
	if err != nil {
		return
	}

	fd, err := os.Open(filename)
	if err != nil {
		return
	}
	lines = make([]string, 0)
	reader := bufio.NewReader(fd)
	prefix := ""
	var isLongLine bool
	for {
		byteLine, isPrefix, er := reader.ReadLine()
		if er != nil && er != io.EOF {
			return nil, er
		}
		if er == io.EOF {
			break
		}
		line := string(byteLine)
		if isPrefix {
			prefix += line
			continue
		} else {
			isLongLine = true
		}

		line = prefix + line
		if isLongLine {
			prefix = ""
		}
		if re.MatchString(line) {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func WriteToFile(file, info string) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("open file %s failed %s", file, err.Error())
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(info))
	if err != nil {
		log.Errorf("write file %s failed %s", file, err.Error())
		return err
	}
	return nil
}

func FirstLine(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return firstLine(file)
}

func LastLine(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return lastLine(filename, file)
}

func firstLine(file *os.File) (string, error) {
	var first []byte
	var offset int64
	for {
		buf := make([]byte, bufSize)
		n, err := file.ReadAt(buf, offset)
		if err != nil && err != io.EOF {
			return "", err
		}

		for i := 0; i < n; i++ {
			if buf[i] == '\n' {
				return string(append(first, buf[:i]...)), nil
			}
		}

		first = append(first, buf[:n]...)
		offset += bufSize
	}
}

func lastLine(filename string, file *os.File) (string, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return "", err
	}

	var last []byte
	offset := info.Size()
	for {
		offset -= bufSize
		if offset < 0 {
			offset = 0
		}
		buf := make([]byte, bufSize)
		n, err := file.ReadAt(buf, offset)
		if err != nil && err != io.EOF {
			return "", err
		}

		if buf[n-1] == '\n' {
			buf = buf[:n-1]
			n--
		} else {
			buf = buf[:n]
		}
		for n--; n >= 0; n-- {
			if buf[n] == '\n' {
				return string(append(buf[n+1:], last...)), nil
			}
		}

		last = append(buf, last...)
	}
}

func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// SelfPath gets compiled executable file absolute path
func SelfPath() string {
	path, _ := filepath.Abs(os.Args[0])
	return path
}

// SelfDir gets compiled executable file directory
func SelfDir() string {
	return filepath.Dir(SelfPath())
}

// GetFileNameWithoutExt 返回没有后缀的文件名，
// 如abc.txt => abc
func GetFileNameWithoutExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func ReadAllIntoMemory(filename string) (content []byte, err error) {
	fp, err := os.Open(filename) // 获取文件指针
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	fileInfo, err := fp.Stat()
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, fileInfo.Size())
	_, err = fp.Read(buffer) // 文件内容读取到buffer中
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

// ReWriteFile用content重新写文件filename
func ReWriteFile(filename, content string) error {
	// 判断文件是否存在
	// exist, err := PathExists(filename)
	// if err != nil {
	// 	return nil
	// }

	// if exist {
	// 	// 判断bak文件是否存在，如果不存在则备份
	// 	bakFile := fmt.Sprintf("%s.bak", filename)
	// 	var bakExist bool
	// 	bakExist, err = PathExists(bakFile)
	// 	if err != nil {
	// 		return err
	// 	} else if !bakExist {
	// 		_, err = Copy(filename, bakFile)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	tmpfile, err := TempFilenameWithText(content)
	if err != nil {
		return err
	}
	defer Unlink(tmpfile)

	_, err = Copy(tmpfile, filename)
	return err
}

func CreateFile(filePath string) {
	var _, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		var file, err = os.Create(filePath)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer file.Close()
	}
	fmt.Println("===done creating file", filePath)
}

func GetRootPath() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

func ReadFile() {
	path, err := os.Getwd()
	file, err := os.Open(path + "\\pkg\\file\\test.txt")
	if err != nil {
		fmt.Printf("There occurred an err on opening the file:%v", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		fmt.Println("The input was:", line)
		if err == io.EOF {
			return
		}
	}
}

func IoutilRead() {
	path, err := os.Getwd()
	filePath := path + "\\pkg\\file\\test.txt"
	buffer, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "File Error: %s\n", err)
		// panic(err.Error())
	}
	fmt.Println(buffer)
	err = ioutil.WriteFile(filePath, buffer, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func WriteFile() {
	path, err := os.Getwd()
	filePath := path + "\\pkg\\file\\test.txt"
	outputFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		return
	}
	defer outputFile.Close()
	writer := bufio.NewWriter(outputFile)
	str := "hello world\n"
	for i := 0; i < 5; {
		i++
		writer.WriteString(str)
	}
	writer.Flush()
}

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		return
	}
	defer dst.Close()

	return io.Copy(dst, src)
}

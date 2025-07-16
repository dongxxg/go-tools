package logger

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LOG_LEVEL int

const (
	ERROR LOG_LEVEL = iota
	CERROR
	WARN
	INFO
	DEBUG
)

var lg = Logger{level: DEBUG}

func sourceDir(file string) string {
	dir := filepath.Dir(file)
	dir = filepath.Dir(dir)

	s := filepath.Dir(dir)
	if filepath.Base(s) != "gorm.io" {
		s = dir
	}
	return filepath.ToSlash(s) + "/"
}

var logSourceDir string

func init() {
	lg.out = lg.outTerminalWithGoId
	lg.ttyWriter = os.Stdout
	_, file, _, _ := runtime.Caller(0)
	// compatible solution to get gorm source directory with various operating systems
	logSourceDir = sourceDir(file)
}

func SetLogFile(logDir string, filename string, buffSize int) {
	lg.SetLogFile(logDir, filename, buffSize)
}
func Debug(format string, v ...any) {
	lg.Debug(format, v...)
}
func Info(format string, v ...any) {
	lg.Info(format, v...)
}
func Warn(format string, v ...any) {
	lg.Warn(format, v...)
}
func Error(format string, v ...any) {
	lg.Error(format, v...)
}
func CError(format string, v ...any) {
	lg.CError(format, v...)
}
func Fatal(format string, v ...any) {
	lg.CError(format, v...)
	os.Exit(1)
}
func PanicLog(errMsg interface{}, errStack []byte) string {
	errStackMsg := string(errStack)
	errStackMsg = strings.Replace(errStackMsg, "\n", "##", -1)
	return fmt.Sprintf("PANIC :%s; PANIC STACK :%s", errMsg, errStackMsg)
}

func outTerminal(header string, format string, v ...any) {
	log.Print(header + fmt.Sprintf(format, v...))
}

func SetLogLevel(lv LOG_LEVEL) {
	lg.SetLogLevel(lv)
}

type Logger struct {
	level      LOG_LEVEL
	fileName   []byte
	dayBegin   int64
	file       *os.File
	fileWriter io.Writer
	ttyWriter  io.Writer // 输出到终端的writer
	bufSize    int
	mu         sync.Mutex
	out        func(header string, format string, v ...any)
	writer     io.Writer // 最终的writer
}

func (l *Logger) Debug(format string, v ...any) {
	if l.level >= DEBUG {
		format = format + " " + fileWithLineNum()
		l.out("DEBUG ", format, v...)
		// l.out("DEBUG  "+fileWithLineNum(), format, v...)
	}
}
func (l *Logger) Info(format string, v ...any) {
	if l.level >= INFO {
		format = format + " " + fileWithLineNum()
		l.out("INFO ", format, v...)
		// l.out("INFO   "+fileWithLineNum(), format, v...)
	}
}
func (l *Logger) Warn(format string, v ...any) {
	if l.level >= WARN {
		format = format + " " + fileWithLineNum()
		l.out("WARN ", format, v...)
		// l.out("WARN   "+fileWithLineNum(), format, v...)
	}
}
func (l *Logger) Error(format string, v ...any) {
	if l.level >= ERROR {
		format = format + " " + fileWithLineNum()
		l.out("ERROR ", format, v...)
		// l.out("ERROR  "+fileWithLineNum(), format, v...)
	}
}
func (l *Logger) CError(format string, v ...any) {
	if l.level >= CERROR {
		format = format + " " + fileWithLineNum()
		l.out("WARN [CLIENT_WARN] ", format, v...)
		// l.out("CERROR "+fileWithLineNum(), format, v...)
	}
}

// 输出到控制台不会经过content通道，可能会乱序
// 加锁防止乱序
func (l *Logger) outTerminal(header string, format string, v ...any) {
	now := time.Now()
	l.mu.Lock()
	l.writer = l.ttyWriter
	defer l.mu.Unlock()

	var buffer bytes.Buffer
	buffer.WriteString(header)
	buffer.Write(now.AppendFormat([]byte{}, "[2006-01-02 15:04:05.000] "))
	buffer.WriteString(fmt.Sprintf(format, v...))
	buf := buffer.Bytes()
	if buf[len(buf)-1] != '\n' {
		buf = append(buf, '\n')
	}
	l.writer.Write(buf)

	// l.writeCore(header+fmt.Sprintf(format, v...), now)
}
func (l *Logger) outTerminalWithGoId(header string, format string, v ...any) {
	header += strconv.FormatUint(CurGoroutineID(), 10) + " "
	l.outTerminal(header, format, v...)
}
func (l *Logger) outFile(header string, format string, v ...any) {
	now := time.Now()
	l.mu.Lock()
	l.writer = l.getFile(now)
	l.mu.Unlock()
	l.writeCore(header+fmt.Sprintf(format, v...), now)
}
func (l *Logger) outFileWithGoId(header string, format string, v ...any) {
	header += strconv.FormatUint(CurGoroutineID(), 10) + " "
	l.outFile(header, format, v...)
}

func (l *Logger) getFile(now time.Time) io.Writer {
	for now.Unix()-l.dayBegin >= 24*3600 {
		l.dayBegin += 24 * 3600
		if l.file != nil {
			l.file.Close()
		}
		l.file = nil
	}
	if l.file == nil {
		var err error
		name := now.AppendFormat(l.fileName, "2006-01-02.log")
		l.file, err = os.OpenFile(string(name), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("%v\n", err)
			return nil
		}
		if l.bufSize != 0 {
			l.fileWriter = bufio.NewWriterSize(l.file, l.bufSize)
		} else {
			l.fileWriter = l.file
		}
	}
	return l.fileWriter
}

func (l *Logger) setTTY() {
	if l.bufSize > 0 {
		l.ttyWriter = bufio.NewWriterSize(os.Stdout, l.bufSize)
	} else {
		l.ttyWriter = os.Stdout
	}
}

// 输出到tty不会走content通道，即不会调用writeCore，所以file没有多余
func (l *Logger) writeCore(info string, now time.Time) {
	buf := make([]byte, 15, 1024) // 前面的"[00:00:00.000] "的15位时间固定
	// buf = now.AppendFormat(buf, "[2006-01-02 15:04:05.000] ") //复制时间;
	buf = formatTime(now, buf)
	buf = append(buf, info[:]...) // 先复制WARN等;
	// buf = append(buf, info[6:]...)                            //复制后面的信息;
	if buf[len(buf)-1] != '\n' {
		buf = append(buf, '\n')
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer.Write(buf)
}

// 格式化成"[00:00:00.000] "
func formatTime(now time.Time, buf []byte) []byte {
	if len(buf) < 15 { // 长度不够，什么也不做
		return buf
	}
	h, m, s := now.Clock()
	ms := now.Nanosecond() / 1e6
	buf[0] = '['
	buf[1], buf[2] = formatTwo(h)
	buf[3] = ':'
	buf[4], buf[5] = formatTwo(m)
	buf[6] = ':'
	buf[7], buf[8] = formatTwo(s)
	buf[9] = '.'
	buf[10], buf[11], buf[12] = formatThree(ms)
	buf[13] = ']'
	buf[14] = ' '
	return buf
}

// 格式化一位，两位，三位
func formatOne(t int) byte {
	return byte(t + '0')
}
func formatTwo(t int) (byte, byte) {
	return formatOne(t / 10), formatOne(t % 10)
}
func formatThree(t int) (byte, byte, byte) {
	b1, b2 := formatTwo(t / 10)
	b3 := formatOne(t % 10)
	return b1, b2, b3
}

func fileWithLineNum() string {
	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.HasPrefix(file, logSourceDir) || strings.HasSuffix(file, "_test.go")) {
			return file + ":" + strconv.FormatInt(int64(line), 10) + " "
		}
	}

	return ""
}

// SetLogFile 设置日志初始参数
// filename
//
//	如果为""，输出到outTerminalWithGoId
//	如果为"tty", 输出到outTerminal
//	否则输出到outFile
//
// buffSize在输出到outFile有效
//
//	如果==0，将不使用带缓存的Writer，而是使用File直接Write
func (l *Logger) SetLogFile(logDir string, filename string, buffSize int) {
	l.dayBegin = DayBegin().Unix()
	l.bufSize = buffSize
	// 如果不需要打文件, 请将console 重定向到/dev/null
	switch filename {
	case "":
		l.out = l.outTerminalWithGoId
		l.setTTY()
	case "tty":
		l.out = l.outTerminal
		l.setTTY()
	default:
		l.fileName = []byte(filepath.Join(logDir, filename))
		_ = os.MkdirAll(logDir, os.ModePerm)
		l.out = l.outFile
	}
}

func (l *Logger) End() {
	time.Sleep(2 * time.Second)
}

func (l *Logger) SetLogLevel(logLevel LOG_LEVEL) {
	l.level = logLevel
}

func DayBegin() time.Time {
	now := time.Now()
	_, offset := now.Zone()
	s := now.Unix()
	s = s - (s+int64(offset))%(24*60*60)
	return time.Unix(s, 0)
}

func GetLogger() *Logger {
	return &lg
}

// Package log 日志模块，依赖于conf模块
package log

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dongxxg/go-tools/common"
	conf "github.com/dongxxg/go-tools/config"
	"github.com/sirupsen/logrus"
)

var log = Noop()

func init() {
	Init()
}

// Init 根据配置初始化日志模块。配置参数：Level - 日志级别，IsJSON - 输出是否以JSON格式，
// RoteteCount - 保留日志文件天数， Path - 日志输出目录
func Init() {
	log = NewLog("")
}

func WithFields(fields Fields) Modular {
	return log.WithFieldz(fields)
}

func WithDuration(d time.Duration) Modular {
	return WithFields(Fields{
		"During": fmt.Sprintf("%.1fms", float32(d)/float32(time.Millisecond)),
	})
}

func SetOutput(stream io.Writer) {
	log = New(stream, NewConfig())
}

// func SetLevel(level string) {
// 	log.SetLevel(level)
// }

// func GetLevel() string {
// 	return log.GetLevel()
// }

func Debug(args ...interface{})                 { log.Debug(args...) }
func Debugln(args ...interface{})               { log.Debugln(args...) }
func Debugf(format string, args ...interface{}) { log.Debugf(format, args...) }

func Error(args ...interface{})                 { log.Error(args...) }
func Errorln(args ...interface{})               { log.Errorln(args...) }
func Errorf(format string, args ...interface{}) { log.Errorf(format, args...) }

func Fatal(args ...interface{})                 { log.Fatal(args...) }
func Fatalln(args ...interface{})               { log.Fatalln(args...) }
func Fatalf(format string, args ...interface{}) { log.Fatalf(format, args...) }

func Info(args ...interface{})                 { log.Info(args...) }
func Infoln(args ...interface{})               { log.Infoln(args...) }
func Infof(format string, args ...interface{}) { log.Infof(format, args...) }

func Trace(args ...interface{})                 { log.Trace(args...) }
func Traceln(args ...interface{})               { log.Traceln(args...) }
func Tracef(format string, args ...interface{}) { log.Tracef(format, args...) }

func Warn(args ...interface{})                 { log.Warn(args...) }
func Warnln(args ...interface{})               { log.Warnln(args...) }
func Warnf(format string, args ...interface{}) { log.Warnf(format, args...) }

func Print(args ...interface{})                 { log.Print(args...) }
func Println(args ...interface{})               { log.Println(args...) }
func Printf(format string, args ...interface{}) { log.Printf(format, args...) }

func SetLevel(level Level) { log.Logger.SetLevel(logrus.Level(level)) }
func GetLevel() Level      { return Level(log.Logger.GetLevel()) }

func AddHook(hook logrus.Hook) { log.Logger.AddHook(hook) }
func NewLog(logname string) *Logger {

	var config Config
	// ut scenario
	if common.WithinUT() || conf.GetConfFile() == "" {
		config = NewConfig()
	} else {
		var localTime bool
		if conf.GetBool("Logging.LocalTime") {
			localTime = false
		} else {
			localTime = true
		}
		config = Config{
			Prefix:       "",
			LogLevel:     conf.GetString("Logging.Level"),
			AddTimeStamp: true,
			JSONFormat:   conf.GetBool("Logging.IsJSON"),
			RotateCnt:    uint(conf.GetInt("Logging.RotateCount")),
			StaticFields: Fields{},
			LogPath:      conf.GetString("Logging.Path"),
			LogfileName:  logname,
			MaxSize:      conf.GetInt("Logging.MaxSize"), // megabytes
			MaxBackups:   conf.GetInt("Logging.MaxBackups"),
			MaxAge:       conf.GetInt("Logging.MaxAge"),    //days
			Compress:     conf.GetBool("Logging.Compress"), // disabled by default
			LocalTime:    localTime,
		}
		if logname != "" {
			config.StaticFields["module"] = logname
		}

		if config.RotateCnt == 0 {
			config.RotateCnt = 7 // 一周
		}
	}

	newLog := New(os.Stdout, config)
	return newLog
}

// KitLog returns a wrapper for go-kit log.Logger
// func KitLog() kitlog.Logger {
// 	return kitlogrus.NewLogger(log)
// }

func GetLogger() *Logger {
	return log
}
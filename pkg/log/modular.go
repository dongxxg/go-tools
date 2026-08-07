package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"

	"unitechs.com/unios-dice/uni-base/core/common"
	conf "unitechs.com/unios-dice/uni-base/core/config"
)

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// Level type
type Level uint32

// Logger level constants
const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

// Logger defines unios logger structure
type Logger struct {
	*logrus.Entry
	config Config
	// stream io.Writer
}

// Config defines the log configuration
type Config struct {
	Prefix       string
	LogLevel     string
	AddTimeStamp bool
	JSONFormat   bool
	StaticFields Fields
	LogPath      string
	RotateCnt    uint
	LogfileName  string
	MaxSize      int
	MaxBackups   int
	MaxAge       int
	Compress     bool
	LocalTime    bool
}

// NewConfig returns a config struct with the default values for each field.
func NewConfig() Config {
	return Config{
		Prefix:       "",
		LogLevel:     "DEBUG",
		AddTimeStamp: true,
		JSONFormat:   true,
		StaticFields: Fields{},
		LogPath:      "",
		RotateCnt:    0,
	}
}

// Noop creates and returns a new logger object that writes nothing.
func Noop() *Logger {
	// caller := ""
	// _, file, no, ok := runtime.Caller(1)
	// if ok {
	// 	caller = fmt.Sprintf("file: %s, line: %d", file, no)
	// }
	// fmt.Printf("Warn: %s using Noop log, print to console only\n", caller)

	return New(os.Stdout, NewConfig())
}

// Noop creates and returns a new logger object that writes nothing.
func ProcessorNoop() Modular {

	return New(os.Stdout, NewConfig())
}
func (l *Logger) WithFields(fields map[string]string) Modular {
	newFields := make(logrus.Fields, len(fields))
	for k, v := range fields {
		newFields[k] = v
	}

	newLogger := *l
	newLogger.Entry = l.Entry.WithFields(newFields)
	return &newLogger
}

// With returns a copy of the logger with new labels added to the logging
// context.
func (l *Logger) With(keyValues ...interface{}) Modular {
	newEntry := l.Entry.WithFields(logrus.Fields{})
	for i := 0; i < (len(keyValues) - 1); i += 2 {
		key, ok := keyValues[i].(string)
		if !ok {
			continue
		}
		newEntry = newEntry.WithField(key, keyValues[i+1])
	}

	newLogger := *l
	newLogger.Entry = newEntry
	return &newLogger
}

// New creates unios logger instance
func New(stream io.Writer, config Config) *Logger {
	l, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		l = logrus.InfoLevel
	}

	logger := logrus.New()
	logger.SetLevel(l)
	//logger.SetOutput(stream)

	isJSON := config.JSONFormat
	if isJSON {
		logger.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp:  !config.AddTimeStamp,
			DisableHTMLEscape: true,
			FieldMap: logrus.FieldMap{
				"time":  "@time",
				"level": "@level",
			},
			TimestampFormat: "2006-01-02T15:04:05.99Z07:00",
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableColors:    true,
			DisableTimestamp: !config.AddTimeStamp,
			FullTimestamp:    true,
			DisableQuote:     true,
		})
	}
	//是否需要输出日志
	if needLocalLogFile() {
		logfilename := getLogfilename(config)
		max := config.MaxAge
		if max == 0 {
			max = int(config.RotateCnt)
		}
		logout := &lumberjack.Logger{
			Filename:   logfilename,
			MaxSize:    config.MaxSize, // megabytes
			MaxBackups: config.MaxBackups,
			MaxAge:     max,             //days
			Compress:   config.Compress, // disabled by default
			LocalTime:  config.LocalTime,
		}
		writers := []io.Writer{logout, stream}
		filewriter := io.MultiWriter(writers...)
		logger.SetOutput(filewriter)
	} else {
		logger.SetOutput(stream)
	}

	fields := logrus.Fields{}
	for k, v := range config.StaticFields {
		fields[k] = v
	}

	return &Logger{
		logger.WithFields(fields),
		config,
	}
}
func getLogfilename(config Config) string {
	binaryFullPath, err := os.Executable()
	if err != nil {
		panic("fatal error getting current binary path")
	}

	var logFullPath string
	if config.LogPath != "" {
		logFullPath = config.LogPath
	} else {
		// follow project layout convention
		logFullPath = filepath.Join(filepath.Dir(binaryFullPath), "logs")
	}

	err = common.MakeDir(logFullPath)
	if err != nil {
		panic(err)
	}

	conf.Set("Logging.Path", logFullPath)

	var tmpName string
	if config.LogfileName == "" {
		currentBinaryName := filepath.Base(binaryFullPath)
		tmpName = strings.ReplaceAll(conf.GetString("Service.Name"), " ", "_")
		if tmpName == "" {
			tmpName = currentBinaryName[:len(currentBinaryName)-len(filepath.Ext(currentBinaryName))]
		}
	} else {
		tmpName = config.LogfileName
	}

	logFileName := fmt.Sprintf("%s.log", tmpName)
	logFile := filepath.Join(logFullPath, logFileName)
	return logFile
}

func newLfsHook(logName string, rotationTime time.Duration, cnt uint) logrus.Hook {
	writer, err := rotatelogs.New(
		logName+".%Y%m%d",
		rotatelogs.WithLinkName(logName),
		rotatelogs.WithRotationTime(rotationTime),

		// WithMaxAge和WithRotationCount只能选其一
		// 如果设置为0，则不进行删除
		rotatelogs.WithRotationCount(cnt),
	)
	if err != nil {
		panic(err)
	}

	isjson := conf.GetBool("Logging.isJSON")
	var formatter logrus.Formatter
	if isjson {
		formatter = &logrus.JSONFormatter{
			DisableHTMLEscape: true,
			FieldMap: logrus.FieldMap{
				"time":  "@time",
				"level": "@level",
			},
			TimestampFormat: time.RFC3339,
		}
	} else {
		formatter = &logrus.TextFormatter{
			DisableColors:   true,
			FullTimestamp:   true,
			DisableQuote:    true,
			TimestampFormat: "2006-01-02T15:04:05.99Z07:00",
		}
	}
	// 可设置按不同level创建不同的文件名，目前都写到一起
	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.TraceLevel: writer,
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, formatter)

	return lfsHook
}

// NewModule is used to log a submodule
func (l *Logger) NewModule(prefix string) Modular {
	config := l.config
	config.Prefix = fmt.Sprintf("%v%v", config.Prefix, prefix)

	return &Logger{
		l.WithField("@module", config.Prefix),
		config,
	}
}

// WithFieldz is samilar with logrus WithFields，just to adapt modular interface
func (l *Logger) WithFieldz(fields Fields) Modular {
	newConfig := l.config
	logrusFields := make(logrus.Fields)
	for k, v := range fields {
		logrusFields[k] = v
	}

	return &Logger{
		l.Entry.WithFields(logrusFields),
		newConfig,
	}
}

func needLocalLogFile() bool {
	if conf.GetConfFile() == "" {
		return false
	}
	// UT时不产生日志
	if common.WithinUT() {
		return false
	}
	// 程序运行在容器中
	if common.IsInContainer() {
		// 检查是否要求容器中生成日志文件，默认容器中应该不产生文件
		return conf.GetBool("Logging.KeepInContainer")
	}
	// 非容器环境，输出日志文件
	return true
}

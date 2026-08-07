package log

//------------------------------------------------------------------------------

// Modular 是日志模块的接口定义，日志模块实现了里面全部接口。
type Modular interface {
	WithFields(fields map[string]string) Modular
	With(keyValues ...interface{}) Modular
	NewModule(prefix string) Modular

	WithFieldz(fields Fields) Modular

	Fatalf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Tracef(format string, v ...interface{})

	Fatalln(args ...interface{})
	Errorln(args ...interface{})
	Warnln(args ...interface{})
	Infoln(args ...interface{})
	Debugln(args ...interface{})
	Traceln(args ...interface{})

	Fatal(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
	Info(args ...interface{})
	Debug(args ...interface{})
	Trace(args ...interface{})
}

package logger

type LogPrefix string

type logFunc func(format string, v ...any)

func (l LogPrefix) Debug(format string, v ...any) {
	l.log(lg.Debug, format, v)
}

func (l LogPrefix) Info(format string, v ...any) {
	l.log(lg.Info, format, v)
}

func (l LogPrefix) Warn(format string, v ...any) {
	l.log(lg.Warn, format, v)
}

func (l LogPrefix) CError(format string, v ...any) {
	l.log(lg.CError, format, v)
}

func (l LogPrefix) Error(format string, v ...any) {
	l.log(lg.Error, format, v)
}

func (l LogPrefix) log(lf logFunc, format string, v []any) {
	if l != "" {
		format = string(l) + " " + format
	}
	lf(format, v...)
}

func (l LogPrefix) Append(ap string) LogPrefix {
	return l + LogPrefix(" ["+ap+"]")
}

package logger

func Append(l ILog, ap string) ILog {
	switch l.(type) {
	case LogPrefix:
		return l.(LogPrefix).Append(ap)
	default:
		return l
	}
}

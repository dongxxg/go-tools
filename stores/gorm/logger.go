package db

import (
	"context"
	"errors"
	"time"

	glog "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	"github.com/dongxxg/go-tools/log"
)

// gormLogger 将 gorm 的日志输出桥接到项目 log 模块(logrus),
// 替代 gorm 内置 logger(直接写标准错误, 脱离统一的日志体系)
type gormLogger struct {
	logLevel      glog.LogLevel
	slowThreshold time.Duration // 慢SQL阈值, 超过按 Warn 级别记录
}

// 编译期保证实现了 gorm 的日志接口
var _ glog.Interface = (*gormLogger)(nil)

func newGormLogger(logLevel glog.LogLevel, slowThreshold time.Duration) *gormLogger {
	return &gormLogger{
		logLevel:      logLevel,
		slowThreshold: slowThreshold,
	}
}

// LogMode 实现 gorm logger 接口, 返回切换了日志级别的新实例
func (l *gormLogger) LogMode(level glog.LogLevel) glog.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= glog.Info {
		log.Infof("%s [gorm] "+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= glog.Warn {
		log.Warnf("%s [gorm] "+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= glog.Error {
		log.Errorf("%s [gorm] "+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Trace 记录每条SQL的执行情况: 出错按 Error, 慢SQL按 Warn, 普通SQL按 Info
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= glog.Silent {
		return
	}

	elapsed := time.Since(begin)
	caller := utils.FileWithLineNum()
	switch {
	// ErrRecordNotFound 是正常业务场景, 不按错误级别记录
	case err != nil && !errors.Is(err, glog.ErrRecordNotFound) && l.logLevel >= glog.Error:
		sql, rows := fc()
		log.Errorf("%s [rows:%v] [elapsed:%v] [error:%v] %s", caller, rows, elapsed, err, sql)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.logLevel >= glog.Warn:
		sql, rows := fc()
		log.Warnf("%s [rows:%v] [elapsed:%v >= %v] SLOW SQL: %s", caller, rows, elapsed, l.slowThreshold, sql)
	case l.logLevel >= glog.Info:
		sql, rows := fc()
		log.Infof("%s [rows:%v] [elapsed:%v] %s", caller, rows, elapsed, sql)
	}
}

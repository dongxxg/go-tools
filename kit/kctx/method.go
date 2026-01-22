// 任务号    日期     开发者      描述

package kctx

import (
	"context"
	"unitechs.com/unios-dice/uni-base/core/log"
)

var ctx Ctx

func GetCtx() Ctx {
	return ctx
}
func SetCtx(c Ctx) {
	ctx = c
}

// NewCtx new ctx for global
func NewCtx(c context.Context, tid string) {
	ctx = NewCtxWithField2(c, tid)
}

func Debug(args ...interface{})                 { ctx.Debug(args...) }
func Debugln(args ...interface{})               { ctx.Debugln(args...) }
func Debugf(format string, args ...interface{}) { ctx.Debugf(format, args...) }

func Error(args ...interface{})                 { ctx.Error(args...) }
func Errorln(args ...interface{})               { ctx.Errorln(args...) }
func Errorf(format string, args ...interface{}) { ctx.Errorf(format, args...) }

func Fatal(args ...interface{})                 { ctx.Fatal(args...) }
func Fatalln(args ...interface{})               { ctx.Fatalln(args...) }
func Fatalf(format string, args ...interface{}) { ctx.Fatalf(format, args...) }

func Info(args ...interface{})                 { ctx.Info(args...) }
func Infoln(args ...interface{})               { ctx.Infoln(args...) }
func Infof(format string, args ...interface{}) { ctx.Infof(format, args...) }

func Trace(args ...interface{})                 { ctx.Trace(args...) }
func Traceln(args ...interface{})               { ctx.Traceln(args...) }
func Tracef(format string, args ...interface{}) { ctx.Tracef(format, args...) }

func Warn(args ...interface{})                 { ctx.Warn(args...) }
func Warnln(args ...interface{})               { ctx.Warnln(args...) }
func Warnf(format string, args ...interface{}) { ctx.Warnf(format, args...) }

func Print(args ...interface{})                 { ctx.Print(args...) }
func Println(args ...interface{})               { ctx.Println(args...) }
func Printf(format string, args ...interface{}) { ctx.Printf(format, args...) }

func GetLevel() log.Level {
	return log.GetLevel()
}

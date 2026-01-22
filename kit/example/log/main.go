package main

import (
	logk "unitechs.com/cno/kit/component-base/log2kafka"
)

func main() {
	//ctx, err := logk.NewCtxWithKafka(getter)
	ctx := logk.NewDefaultContext()
	ctx.Infof("testInfo")
	ctx.Infof("testInfo2")
	ctx.Infof("testInfo3")
	select {}
}

package cron

import (
	"github.com/robfig/cron"
	"log"
	"time"
)

func Cron() {
	log.Println("Starting...")

	// 会根据本地时间创建一个新（空白）的 Cron job runner
	c := cron.New()
	// AddFunc会向Cron job runner添加一个func，以按给定的时间表运行，会首先解析时间表，
	//如果填写有问题会直接 err，无误则将 func 添加到 Schedule 队列中等待执行
	c.AddFunc("* * * * * *", func() {
		log.Println("Run...")
	})
	c.Start() // 在当前执行的程序中启动 Cron 调度程序。

	t1 := time.NewTimer(time.Second * 10) //会创建一个新的定时器，持续你设定的时间d后发送一个channel消息
	for {                                 //阻塞 select 等待 channel
		select {
		case <-t1.C:
			t1.Reset(time.Second * 10) // 会重置定时器，让它重新开始计时
		}
	}
}

func Cron2() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		<-ticker.C
	}
	ticker.Stop()
}

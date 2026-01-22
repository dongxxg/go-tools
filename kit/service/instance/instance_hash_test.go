package instance

import (
	"context"
	"testing"
	"time"
	"unitechs.com/cno/kit/db/kredis"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/util"
)

func TestInstanceHash(t *testing.T) {
	//name := "insA"
	ctx := kctx.NewContext(context.TODO(), "aaaaaaaaaaa")
	//rds, err := kredis.NewRedisRepo(ctx, "node", "localhost:6379", "wiken", 0)
	rds, err := kredis.NewRedisRepo(ctx, "node", "127.0.0.1:6379", "wiken", 0)
	if err != nil {
		panic(err)
	}

	insA := InitInstanceHash(ctx, rds, "unios-alarm:alarm:instanceMap:1", 2, 5, 3)

	insA.IsMaster()
	time.Sleep(time.Second * 10)

	//for i := 0; i < 1000; i++ {
	//	alarmId := util.NewUUIDStr()
	//	isDeal := insA.IsDeal(ctx, alarmId)
	//	master := IsMaster()
	//	fmt.Printf("%s %s sumAlarmId: %s, insNum: %d, insAll: %d, isDeal: %v, isMaster %v",
	//		time.Now().Format(time.RFC3339),
	//		name,
	//		alarmId,
	//		insA.instanceNum,
	//		insA.allNumOfInstance,
	//		isDeal,
	//		master,
	//	)
	//}

}
func BenchmarkName(b *testing.B) {
	name := "insA"
	ctx := kctx.NewContext(context.TODO(), "aaaaaaaaaaa")
	rds, err := kredis.NewRedisRepo(ctx, "node", "localhost:6379", "wiken", 0)
	if err != nil {
		panic(err)
	}

	insA := InitInstanceHash(ctx, rds, "unios-alarm:alarm:instanceMap", 2, 5, 3)
	insA.Run()

	insA.IsMaster()

	for i := 0; i < b.N; i++ {
		alarmId := util.NewUUIDStr()
		isDeal := insA.IsDeal(ctx, alarmId)
		master := IsMaster()
		b.Logf("%s sumAlarmId: %s, insNum: %d, insAll: %d, isDeal: %v, isMaster %v",
			name,
			alarmId,
			insA.instanceNum,
			insA.allNumOfInstance,
			isDeal,
			master,
		)
	}
}

func TestInstanceHashB(t *testing.T) {
	name := "insB"
	ctx := kctx.NewContext(context.TODO(), "aaaaaaaaaaa")
	rds, err := kredis.NewRedisRepo(ctx, "node", "localhost:6379", "wiken", 0)
	if err != nil {
		panic(err)
	}

	insA := InitInstanceHash(ctx, rds, "unios-alarm:alarm:instanceMap", 2, 5, 3)
	insA.Run()

	insA.IsMaster()

	go func() {
		for {
			alarmId := util.NewUUIDStr()
			isDeal := insA.IsDeal(ctx, alarmId)
			master := insA.IsMaster()
			t.Logf("%s sumAlarmId: %s, insNum: %d, insAll: %d, isDeal: %v isMaster %v",
				name,
				alarmId,
				insA.instanceNum,
				insA.allNumOfInstance,
				isDeal,
				master,
			)
			time.Sleep(time.Second)
		}
	}()
	select {}
}
func TestInstanceHashC(t *testing.T) {
	name := "insC"
	ctx := kctx.NewContext(context.TODO(), "aaaaaaaaaaa")
	rds, err := kredis.NewRedisRepo(ctx, "node", "localhost:6379", "wiken", 0)
	if err != nil {
		panic(err)
	}

	insA := NewInstanceHash(ctx, rds, "unios-alarm:alarm:instanceMap", 1, 1, 3)
	insA.Run()

	insA.IsMaster()

	go func() {
		for {
			alarmId := util.NewUUIDStr()
			isDeal := insA.IsDeal(ctx, alarmId)
			master := insA.IsMaster()
			t.Logf("%s sumAlarmId: %s, insNum: %d, insAll: %d, isDeal: %v isMaster %v",
				name,
				alarmId,
				insA.instanceNum,
				insA.allNumOfInstance,
				isDeal,
				master,
			)
			time.Sleep(time.Second)
		}
	}()
	select {}
}

func TestDemo2(t *testing.T) {
	ctx := kctx.NewContext(context.TODO(), "aaaaaaaaaaa")
	rds, err := kredis.NewRedisRepo(ctx, "node", "localhost:6379", "wiken", 0)
	if err != nil {
		panic(err)
	}

	insB := NewInstanceHash(ctx, rds, "unios-alarm:alarm:instanceMap", 2, 5, 3)

	go func() {
		for {
			alarmId := util.NewUUIDStr()
			isDeal := insB.IsDeal(ctx, alarmId)
			t.Logf("insB sumAlarmId: %s, insNum: %d,insAll: %d, isDeal: %v",
				alarmId,
				insB.instanceNum,
				insB.allNumOfInstance,
				isDeal)
			time.Sleep(time.Second)
		}
	}()
	select {}
}

func BenchmarkIsDeal(b *testing.B) {
	ins := NewInstanceHash(kctx.NewDefaultContext(), nil, "", 0, 0, 0)
	ctx := kctx.NewDefaultContext()
	ins.allNumOfInstance = 10
	for i := 0; i < b.N; i++ {
		ins.IsDeal(ctx, util.NewUUIDStr())
	}
}

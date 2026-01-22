package kafka_group

import (
	"testing"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/unios-dice/uni-base/lib/stores/redis"
)

func TestGroupGetter_GetGroup(t *testing.T) {
	ctx := kctx.NewDefaultContext()
	rds := redis.NewRedis("127.0.0.1:6379", "node", "wiken")
	g, err := NewGroupGetter(ctx,
		"unios-alarm:process:kafka-process-demo",
		"unios-alarm:process:kafka-process-demo:lock", rds,
		10*time.Second, 5*time.Second,
	)
	if err != nil {
		t.Fatal(err)
		return
	}
	groups := []string{"group1", "group2", "group3", "group4"}

	g.SetGroups(groups...)
	g.Run(ctx)
	//g.rdsLock.Flush()
	//for {

	for i := 0; i < len(groups); i++ {
		//v := groups[i]
		go func(v string) {
			c := kctx.NewDefaultContext()
			//c.Infof("begin get group: %s", v)
			group, err := g.GetGroup(c, v)
			if err != nil {
				t.Error(err)
				return
			}
			c.Infof("get group: %s", group)
		}(groups[i])
	}
	//time.Sleep(time.Second)
	//}

	select {}
}

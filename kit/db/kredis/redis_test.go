// 任务号    日期     开发者      描述

package kredis

import (
	"context"
	"testing"
	"unitechs.com/cno/kit/kctx"
)

func TestSerialNo(t *testing.T) {
	ctx := kctx.NewContext(context.TODO(), "aaaaaa")
	r, err := NewRedisRepo(ctx, "node", "192.168.5.136:17001", "Jypzcqn!@)!*666", 0)
	if err != nil {
		t.Fatal(err)
	}

	//n, err := r.GenSerialNoWithLua(ctx, "asdf", 100, 1)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Log("serialNo: ", n)

	//for i := 0; i < 10; i++ {
	//	go func() {
	//		for j := 0; j < 1000; j++ {
	//			n, err := r.GenSerialNoWithLua(ctx, "serialNo:test", 100, 1)
	//			if err != nil {
	//				t.Fatal(err)
	//			}
	//			t.Log("serialNo: ", n)
	//		}
	//	}()
	//}
	for i := 0; i < 1000; i++ {
		go func(j int) {
			n, err := r.GenSerialNoWithLua(ctx, "serialNo:test", 100, 1)
			if err != nil {
				t.Error(err)
			}
			t.Logf("%d serialNo: %d", j, n)
		}(i)
	}
	select {}
}

func BenchmarkSerialNo(b *testing.B) {
	ctx := kctx.NewContext(context.TODO(), "aaaaaa")
	r, err := NewRedisRepo(ctx, "node", "192.168.5.136:17001", "Jypzcqn!@)!*666", 0)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		r.GenSerialNoWithLua(ctx, "serialNo:test", 100, 1)
	}
}

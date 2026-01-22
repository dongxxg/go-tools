package otel

import (
	"context"
	"fmt"
	"testing"
	"time"
)

//func TestNew(t *testing.T) {
//	if err := New(kctx.NewDefaultContext(), "http://192.168.5.136:4317", "test"); err != nil {
//		t.Fatal()
//	}
//	_, span := NewSpan(context.TODO(), "test01")
//	span.End()
//	select {}
//}

func TestNew2(t *testing.T) {
	//err := tracerProvider("http://127.0.0.1:14268/api/traces")
	err := NewTraceProvider("http://192.168.5.136:14268/api/traces", "unios-alarm-0511")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.TODO()
	defer Shutdown(ctx)

	spanCtx, span := TraceAndStart(ctx, "", "p3-01")
	defer span.End()

	spanCtx3, span2 := TraceAndStart(ctx, "p4", "p3-02")
	defer span2.End()
	span2.AddEvent("judge")
	// 生成 id 发送给 其他服务
	id := GetTraceParent(spanCtx3)
	fmt.Println("id", id)

	fmt.Println(spanCtx, span2, spanCtx3)
	time.Sleep(time.Second)
}
func TestProcess2(t *testing.T) {
	//err := tracerProvider("http://127.0.0.1:14268/api/traces")
	err := NewTraceProvider("http://192.168.5.136:14268/api/traces", "unios-alarm-process")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.TODO()
	defer Shutdown(ctx)
	// 从其他服务传入的id
	id := "00-04c59dbe2b741d6735f55cfb9d140cf3-4b6442d4d51eec59-01"
	ctx, sp := TraceFromId(id, "", "p04-04")
	defer sp.End()
}

package test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/container/gmap"
	"net"
	"regexp"
	"testing"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/server/socket"
	"unitechs.com/cno/kit/lib/server/socket/handler/subscribe"
)

func TestServer(t *testing.T) {
	srv, err := getServ()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			srv.Broadcast(kctx.NewDefaultContext(), []byte("demo;"), 0)
			time.Sleep(time.Second)
		}
	}()
	srv.Serv()
}

func TestClient(t *testing.T) {
	client()
}
func TestName(t *testing.T) {
	client()
}

func client() {
	ccc := kctx.NewDefaultContext()
	conn, err := net.Dial("tcp", "127.0.0.1:18889")
	if err != nil {
		ccc.Fatal(err)
	}

	defer conn.Close()

	sh := getSubHandler()

	go func() {
		ctx := kctx.NewDefaultContext()
		for {
			msg := sh.GetLoginMsg()
			header := sh.GetHeaderByType(ctx, subscribe.MsgTypeReqLoginAlarm, len(msg))
			header = append(header, []byte(msg)...)
			if _, err := conn.Write(header); err != nil {
				ctx.Errorf("login failed: %v", err)
			} else {
				ctx.Error("login success")
			}

			time.Sleep(time.Second * 1000)
		}
	}()
	go func() {
		ctx := kctx.NewDefaultContext()
		for {
			msg := "ackHeartBeat;reqId=100"
			header := sh.GetHeaderByType(ctx, subscribe.MsgTypeReqHeartBeat, len(msg))
			header = append(header, []byte(msg)...)
			if _, err := conn.Write(header); err != nil {
				ctx.Errorf("heartbeat failed: %v", err)
			} else {
				ctx.Error("heartbeat success")
			}
			time.Sleep(time.Second * 50)
		}
	}()
	go func() {
		ctx := kctx.NewDefaultContext()
		buf := bufio.NewReader(conn)

		for {
			b, err := buf.ReadBytes(';')
			if err != nil {
				ctx.Errorf("read failed")
			}
			ctx.Infof("recv info: %s", string(b))
		}
	}()

	select {}
}

func getServ() (*socket.Server, error) {

	sh := getSubHandler()
	hh, err := subscribe.NewHandler(*sh)
	if err != nil {
		return nil, err
	}

	srv, err := socket.NewServerWith(kctx.NewDefaultContext(), "tcp", "127.0.0.1:18889", hh)
	if err != nil {
		return nil, err
	}

	return srv, nil
}
func getSubHandler() *subscribe.Param {
	msgHeader := `
[
    {"name":"startSign", "type": "16", "length": 2, "value": "0xFFFF"},
    {"name":"msgType", "type": "10", "length": 1, "value": ""},
    {"name":"timeStamp", "type": "10", "length": 4, "value": ""},
    {"name":"lenOfBody", "type": "10", "length": 2, "value": ""}
]`
	h := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(msgHeader), &h); err != nil {
		return nil
	}

	param := `
{
"sockMode":"sever",
"loginReqMsg":"reqLoginAlarm;user=${loginUser};key=${loginPass};type=msg",
"msgHeaderFormat":"",
"loginReqMsg":"reqLoginAlarm;user=${loginUserName};key=${loginPassword};type=msg",
"loginUserName":"wiken",
"loginPassword":"wikenPass",
"loginResMsgSucc":"ackLoginAlarm;result=succ;resDesc=",
"loginResMsgFail":"ackLoginAlarm;result=fail;resDesc=username-error",
"loginResMsgType":"2",
"loginFailRetries":"3",
"heartbeatResMsg":"ackHeartBeat;reqId=${seqNoExpr}",
"seqNoExpr":"reqId=(\\d+)",
"heartbeatReqMsgExpr":"ackHeartBeat;reqId=\\d+",
"closeConnReqMsg":"ackHeartBeat;reqId=",
"realTimeAlarmMsgType":"closeConnAlarm",
"realTimeAlarmMsgType":"",
"shakehandFails":""
}
`
	header := `
[

	{"name":"startSign", "type": "16", "length": 2, "value": "0xFFFF"},
	{"name":"msgType", "type": "10", "length": 1, "value": ""},
	{"name":"timeStamp", "type": "10", "length": 4, "value": ""},
	{"name":"lenOfBody", "type": "10", "length": 2, "value": ""}

]
`

	pp := gmap.NewStrStrMap(true)
	if err := pp.UnmarshalJSON([]byte(param)); err != nil {
		return nil
	}
	m := make([]map[string]interface{}, 0)
	json.Unmarshal([]byte(header), &m)
	b, _ := json.Marshal(m)
	pp.Set("msgHeaderFormat", string(b))
	return subscribe.NewParam(pp, func(key, val string) string {
		return val
	})
}

func TestDemo(t *testing.T) {
	a := "ackHeartBeat;reqId=100"
	r := regexp.MustCompile("regId=(\\d+?)")
	lst := r.FindStringSubmatch(a)
	fmt.Println(lst)

}

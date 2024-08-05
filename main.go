package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Person struct {
	Name    string
	Age     string
	Address string
}

func F(n int) func() int {
	return func() int {
		n++
		return n
	}
}

var f = F(2)

type PublishMsg struct {
	AzCode        string
	ResourceId    string
	ResourceType  string
	ResourceState string
	FixedIps      []string //nova instance fixed_ips
}

func parseFixedIps(payload map[string]interface{}, msg *PublishMsg) {
	if fixedIps, exist := payload["fixed_ips"]; exist {
		var resFixedIps []string
		for _, fixedIp := range fixedIps.([]interface{}) {
			if address, exist := fixedIp.(map[string]interface{})["address"]; exist {
				resFixedIps = append(resFixedIps, address.(string))
			}
		}
		msg.FixedIps = resFixedIps
	}
}

func byteToMap(body []byte) map[string]interface{} {
	msg := make(map[string]interface{})
	err := json.Unmarshal(body, &msg)
	if err != nil {
		fmt.Println(err)
	}
	return msg
}
func watShadowDefer(i int) (ret int) {
	ret = i * 2
	if ret > 10 {
		ret := 10
		defer func() {
			ret = ret + 1
		}()
	}
	return
}

type RoundRobin struct {
	indexer map[string]*uint32
}

func (r *RoundRobin) Select(topic string) uint32 {
	id, exist := r.indexer[topic]
	if !exist {
		var index uint32 = 0
		id = &index
		r.indexer[topic] = id
	}
	*id++
	return *id % 10
}

const AESKEY = "KqxqU4"

// DecodeString decode ecs user to openstack password
func DecodeString(src string) string {
	dst, _ := hex.DecodeString(src)
	dst = dst[:len(dst)-len(AESKEY)]
	return string(dst)
}

func main() {
	fmt.Println(DecodeString("65386561323366342d306339352d343838302d386531372d3361663333386463633230314b7178715534"))

	fmt.Printf("%v", []byte("1233"))
	fmt.Println(time.Now().UnixMicro())
}

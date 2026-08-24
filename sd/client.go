package sd

import (
	"github.com/dongxxg/go-tools/sd/consul"
	"strings"
	"sync"

	"github.com/dongxxg/go-tools/sd/eureka"
	inf "github.com/dongxxg/go-tools/sd/interface"
	"github.com/dongxxg/go-tools/sd/k8s"
	"github.com/dongxxg/go-tools/sd/nacos"
)

var sdClient inf.Client

var once sync.Once

// ClientInstance返回全局的服务注册发现客户端。
// type：k8s,nacos,consul。
// 应用使用前需要检查是否为nil。
func ClientInstance(typ string) inf.Client {
	// k8s环境理论上应该走k8s的服务发现，但是现在上云过渡阶段，很多场景在k8s和Nacos混用
	// 电信的云道目前不支持k8s的headless服务
	once.Do(func() {
		t := strings.ToLower(typ)
		if t == "k8s" {
			sdClient = k8s.NewClient()
		} else if t == "nacos" {
			sdClient = nacos.NewClient()
		} else if t == "consul" { // 没有使用场景
			sdClient = consul.NewClient()
		} else if t == "eureka" {
			sdClient = eureka.NewClient()
		}
	})

	return sdClient
}
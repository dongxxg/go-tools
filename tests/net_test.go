/**
  @Author: dongxx
  @Since: 2026/3/12 15:26
  @Desc: //TODO
**/

package tests

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNet(t *testing.T) {
	ips, err := net.LookupIP("www.baidu.com")
	if err != nil {
		t.Logf("lookupIP failed %s", err.Error())
	}
	t.Logf("ips %v", ips)

	// 使用自定义 DNS 服务器进行解析
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, network, "114.114.114.114:53")
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ips2, err := resolver.LookupIP(ctx, "ip", "www.baidu.com")
	if err != nil {
		t.Logf("lookup ip failed %s", err.Error())
		return
	}
	t.Logf("ips %v", ips2)
}
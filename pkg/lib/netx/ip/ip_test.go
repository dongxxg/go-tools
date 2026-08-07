package ip

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLocalIPbyUDP(t *testing.T) {
	ip := InitLocalIPbyUDP()
	if ip == "" {
		t.Fatal(errors.New("Not work"))
	}
	t.Logf("Pass: %s", ip)
}

func TestValidateIP(t *testing.T) {
	ip1 := "1.1.1.1"
	r1 := ValidateIP(ip1)
	assert.True(t, r1)

	ip2 := "2409::"
	r2 := ValidateIP(ip2)
	assert.True(t, r2)

	ip3 := "10.10.2.300"
	r3 := ValidateIP(ip3)
	assert.False(t, r3)

	ip4 := "10.10.2.100？a=1"
	r4 := ValidateIP(ip4)
	assert.False(t, r4)
}

func TestIsIPv4(t *testing.T) {
	str := "1.1.1.1"
	assert.True(t, IsIPv4(str))
	str = "123423"
	assert.False(t, IsIPv4(str))
}

func TestIsIPv6(t *testing.T) {
	str := "2409:8000::11"
	assert.True(t, IsIPv6(str))
	str = "123423"
	assert.False(t, IsIPv6(str))
}

func TestDualStack(t *testing.T) {
	ips := []string{"192.168.118.220", "2409:8000::1"}
	b, e := IsDualStackIPs(ips)
	assert.Nil(t, e)
	assert.True(t, b)

	ips1 := []string{"1.1.1.1", "2.2.2.2"}
	b, e = IsDualStackIPs(ips1)
	assert.Nil(t, e)
	assert.False(t, b)

	ips2 := []string{"2409::8000::1"}
	b, e = IsDualStackIPs(ips2)
	assert.NotNil(t, e)
}

func TestIsSameAddress(t *testing.T) {
	ip1 := "2408:00E1:0001::1"
	ip2 := "2408:e1:1::1"

	assert.True(t, IsSameAddress(ip1, ip2))

	ip3 := "2408:E1:1::2"
	assert.False(t, IsSameAddress(ip1, ip3))

	ip4 := "192.168.118.220"
	ip5 := "127.0.0.1"
	assert.False(t, IsSameAddress(ip4, ip5))

	ip6 := "fe80::483d:0000:0:4555"
	ip7 := "fe80::483d::4555"
	assert.False(t, IsSameAddress(ip6, ip7))

	ip8 := "::"
	ip9 := "0:0:0:0:0:0:0:0"
	assert.True(t, IsSameAddress(ip8, ip9))

	ip10 := "fe80:483d:0000:0:0:0:0:4555"
	ip11 := "fe80:483d::4555"
	assert.True(t, IsSameAddress(ip10, ip11))
}

func TestIP2CIDRNotation(t *testing.T) {
	ip := "10.21.250.2"
	mask := "255.255.255.0"
	rst := IP2CIDRNotation(ip, mask)
	assert.Equal(t, "10.21.250.2/24", rst)
}

func TestNormalAddress(t *testing.T) {
	ip := "localhost"
	port := 123
	rst := NormalAddress(ip, port)
	assert.Equal(t, "localhost:123", rst)

	ip = "1.2.3.4"
	port = 234
	rst = NormalAddress(ip, port)
	assert.Equal(t, "1.2.3.4:234", rst)

	ip = "fe80::"
	port = 453
	rst = NormalAddress(ip, port)
	assert.Equal(t, "[fe80::]:453", rst)
}

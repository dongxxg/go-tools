package ip

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// var log = logger.NewLogger("gnms", "common/util")

var (
	// default local ip
	localip = "127.0.0.1"

	// default timeout 5 sec
	dialtimeout time.Duration = 5 * time.Second
)

func InitLocalIP(ipaddr string) string {
	if ipaddr != "" {
		addr := strings.Replace(ipaddr, "http://", "", -1)
		conn, err := net.DialTimeout("tcp", addr, dialtimeout)
		if err != nil {
			return localip
		}
		defer conn.Close()
		localip = strings.Split(conn.LocalAddr().String(), ":")[0]
	}

	return localip
}

func InitLocalIPbyUDP() string {
	if localip != "127.0.0.1" {
		return localip
	}

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return localip
	}
	defer conn.Close()

	ip := strings.Split(conn.LocalAddr().String(), ":")[0]
	if ip != "" {
		localip = ip
	}

	return localip
}

// LocalIPs return all non-loopback IPv4 addresses
func LocalIPv4s() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	return ips, nil
}

// LocalIPs return all global unicast IPv6 addresses
func LocalIPv6s() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() && ipnet.IP.To4() == nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	return ips, nil
}

// GetIPv4ByInterface return IPv4 address from a specific interface IPv4 addresses
func GetIPv4ByInterface(name string) ([]string, error) {
	var ips []string

	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	return ips, nil
}

// ipv4,简单计算
func GetIPPair30(ip string) string {
	m := strings.Split(ip, ".")
	i, err := strconv.Atoi(m[len(m)-1])
	if err != nil {
		return ""
	}
	var pos int
	if i%2 == 0 {
		pos = i - 1
	} else {
		pos = i + 1
	}

	a := m[:len(m)-1]
	a = append(a, strconv.Itoa(pos))

	return strings.Join(a, ".")
}

func InternalIP() string {
	infs, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, inf := range infs {
		if isEthDown(inf.Flags) || isLoopback(inf.Flags) {
			continue
		}

		addrs, err := inf.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	return ""
}

func isEthDown(f net.Flags) bool {
	return f&net.FlagUp != net.FlagUp
}

func isLoopback(f net.Flags) bool {
	return f&net.FlagLoopback == net.FlagLoopback
}

// GetIntranetIPArray retrieves and returns the intranet ip list of current machine.
func GetIntranetIPArray() (ips []string, err error) {
	interFaces, e := net.Interfaces()
	if e != nil {
		return ips, e
	}
	for _, interFace := range interFaces {
		if interFace.Flags&net.FlagUp == 0 {
			// interface down
			continue
		}
		if interFace.Flags&net.FlagLoopback != 0 {
			// loopback interface
			continue
		}
		// ignore warden bridge
		if strings.HasPrefix(interFace.Name, "w-") {
			continue
		}
		addresses, e := interFace.Addrs()
		if e != nil {
			return ips, e
		}
		for _, addr := range addresses {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				// not an ipv4 address
				continue
			}
			ipStr := ip.String()
			if IsIntranet(ipStr) {
				ips = append(ips, ipStr)
			}
		}
	}
	return ips, nil
}

// IsIntranet checks and returns whether given ip an intranet ip.
//
// Local: 127.0.0.1
// A: 10.0.0.0--10.255.255.255
// B: 172.16.0.0--172.31.255.255
// C: 192.168.0.0--192.168.255.255
func IsIntranet(ip string) bool {
	if ip == "127.0.0.1" {
		return true
	}
	array := strings.Split(ip, ".")
	if len(array) != 4 {
		return false
	}
	// A
	if array[0] == "10" || (array[0] == "192" && array[1] == "168") {
		return true
	}
	// C
	if array[0] == "192" && array[1] == "168" {
		return true
	}
	// B
	if array[0] == "172" {
		second, err := strconv.ParseInt(array[1], 10, 64)
		if err != nil {
			return false
		}
		if second >= 16 && second <= 31 {
			return true
		}
	}
	return false
}

// GetMacArray retrieves and returns all the mac address of current host.
func GetMacArray() (macs []string, err error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if macAddr == "" {
			continue
		}
		macs = append(macs, macAddr)
	}
	return macs, nil
}

// IP2long converts ip address to an uint32 integer.
func IP2long(ip string) uint32 {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return 0
	}
	return binary.BigEndian.Uint32(netIP.To4())
}

// Long2ip converts an uint32 integer ip address to its string type address.
func Long2ip(long uint32) string {
	ipByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ipByte, long)
	return net.IP(ipByte).String()
}

// Validate checks whether given <ip> a valid IPv4 address.
func ValidateIP(ip string) bool {
	addr := net.ParseIP(ip)
	return addr != nil
}

func IsIPv4(ip string) bool {
	addr := net.ParseIP(ip)
	return addr != nil && addr.To4() != nil
}

// IsIPv6 returns if ip is IPv6.
func IsIPv6(ip string) bool {
	addr := net.ParseIP(ip)
	return addr != nil && addr.To4() == nil
}

// IsDualStackIPs returns if a slice of ips is:
// - all are valid ips
// - at least one ip from each family (v4 or v6)
func IsDualStackIPs(ips []string) (bool, error) {
	v4Found := false
	v6Found := false
	for _, ip := range ips {
		if net.ParseIP(ip) == nil {
			return false, fmt.Errorf("ip %v is invalid", ip)
		}

		if v4Found && v6Found {
			continue
		}

		if IsIPv6(ip) {
			v6Found = true
			continue
		}

		v4Found = true
	}

	return (v4Found && v6Found), nil
}

// 根据ip和端口拼接出适配ipv4和ipv6的字符串
func NormalAddress(ip string, port int) string {
	addr := net.ParseIP(ip)
	if addr == nil {
		// ip是域名或者主机名等
		return fmt.Sprintf("%s:%d", ip, port)
	}

	// ipv4
	if addr.To4() != nil {
		return fmt.Sprintf("%s:%d", ip, port)
	}

	// ipv6
	return fmt.Sprintf("[%s]:%d", ip, port)
}

// IsSameAddress 判断两个IP地址是否相同，主要用于IPv6的各种不同表现形式
func IsSameAddress(ip1, ip2 string) bool {
	aip1 := net.ParseIP(ip1)
	aip2 := net.ParseIP(ip2)
	if aip1 == nil || aip2 == nil {
		return false
	}

	return aip1.Equal(aip2)
}

func IP2CIDRNotation(ip, mask string) string {
	netMask := net.IPMask(net.ParseIP(mask).To4())
	size, _ := netMask.Size()
	return fmt.Sprintf("%s/%d", ip, size)
}

package common

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// core包不应该依赖于lib，因此，一些用到的函数在core实现一遍

var (
	localip = "127.0.0.1"
)

// WithinUT 判断是否在运行testing
func WithinUT() bool {
	if strings.HasSuffix(os.Args[0], ".test") || strings.Contains(os.Args[0], "/_test/") ||
		(len(os.Args) > 1 && (strings.HasSuffix(os.Args[1], ".bench") || // benchmark test
			strings.HasPrefix(os.Args[1], "-test.run"))) { // debug test
		fmt.Println("In UT scenario")
		return true
	}
	return false
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

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func MakeDir(dir string) error {
	exist, err := PathExists(dir)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	err = os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// 是否运行在k8s环境中
func IsInK8S() bool {
	// 在k8s环境中始终会导出这个环境变量
	_, exist := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	return exist
}

// 是否运行在docker环境中
func IsInContainer() bool {
	// 在k8s环境中始终会导出这个环境变量
	_, exist := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	if exist {
		return true
	}

	// 在其它docker环境中检查是否存在.dockerenv文件
	exist, _ = PathExists("/.dockerenv")
	return exist
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

// 读取/.release.version文件数据
func ReadReleaseVersion() (map[string]string, error) {
	filebytes, err := os.ReadFile("./.release.version")
	if err != nil {
		return nil, err
	}
	var versiondata map[string]string
	err = json.Unmarshal(filebytes, &versiondata)
	if err != nil {
		return nil, err
	}
	return versiondata, nil
}

var StartTime time.Time

func GetStartTime() time.Time {
	return StartTime
}

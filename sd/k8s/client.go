package k8s

import (
	// "k8s.io/client-go/kubernetes"
	// typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	// "k8s.io/client-go/rest"

	"fmt"
	"math/rand"
	"net"

	inf "github.com/dongxxg/go-tools/sd/interface"
)

// Client 基于K8s DNS的服务实例发现
// 只支持部分方法，只能返回IP，没有端口
// 如果使用k8s client-go的API，需要在集群中配置对应的role、rolebinding等
type Client struct {
}

func NewClient() inf.Client {
	return &Client{}
}

// Register K8S 不需要
func (c *Client) Register() error {
	return nil
}

// Deregister K8S 不需要
func (c *Client) Deregister() error {
	return nil
}

// 返回服务信息
func (c *Client) Service(serviceName string) (*inf.Service, error) {
	return nil, fmt.Errorf("not support")
}

// 返回一个健康实例
func (c *Client) Instance(serviceName string) (*inf.Instance, error) {
	ips, err := net.LookupIP(serviceName)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("server can't find %s: NXDOMAIN", serviceName)
	} else if len(ips) == 1 {
		return &inf.Instance{
			IP: ips[0].String(),
		}, nil
	}

	index := rand.Intn(len(ips))
	return &inf.Instance{
		IP: ips[index].String(),
	}, nil
}

// 返回一个健康实例的IP:端口，不支持
func (c *Client) InstanceAddress(serviceName string) string {
	return ""
}

// 根据过滤条件返回实例。对于普通service，只会返回一个cluserIP，对于headless service，返回所有pod IP
func (c *Client) Instances(serviceName string, filter map[string]interface{}) ([]inf.Instance, error) {
	ips, err := net.LookupIP(serviceName)
	if err != nil {
		return nil, err
	}

	result := make([]inf.Instance, 0, 10)
	for _, ip := range ips {
		result = append(result, inf.Instance{
			IP: ip.String(),
		})
	}
	return result, nil
}

// func getk8sClient() (typev1.CoreV1Interface, error) {
// 	// creates the in-cluster config
// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		return nil, err
// 	}
// 	// creates the clientset
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return clientset.CoreV1(), nil
// }
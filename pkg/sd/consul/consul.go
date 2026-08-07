package consul

import (
	"fmt"

	stdconsul "github.com/hashicorp/consul/api"
	"unitechs.com/unios-dice/uni-base/core/common"
	"unitechs.com/unios-dice/uni-base/core/config"
	"unitechs.com/unios-dice/uni-base/core/log"
	inf "unitechs.com/unios-dice/uni-base/core/sd/interface"
)

// TODO 未测试
type Client struct {
	consul *stdconsul.Client
}

func NewClient() inf.Client {
	host := config.GetString("SD.Server.IpAddr")
	port := config.GetInt("SD.Server.Port")

	consulConfig := stdconsul.DefaultConfig()
	consulConfig.Address = common.NormalAddress(host, port)
	// consulConfig.Token = token

	stdClient, err := stdconsul.NewClient(consulConfig)
	if err != nil {
		log.Errorf("Failed to create Consul client: %v", err)
		return nil
	}

	return &Client{
		consul: stdClient,
	}
}

// Register 自动向Nacos注册服务，如果配置文件中有IP、Port配置，则使用配置文件中数据
func (c *Client) Register() error {
	name, ip, port, err := inf.CreateRegistion()
	if err != nil {
		return err
	}

	param := &stdconsul.AgentServiceRegistration{
		ID:                fmt.Sprintf("%s-%s-%d", name, ip, port),
		Name:              name,
		Port:              int(port),
		Tags:              []string{},
		Address:           ip,
		EnableTagOverride: false,
		// skipping check(s)
	}
	if err := c.consul.Agent().ServiceRegister(param); err != nil {
		log.Errorf("Register service failed! param:%+v, error: %+v", param, err)
		return err
	}

	log.Infof("register service succeeded, param:%+v", param)
	return nil
}

func (c *Client) Deregister() error {
	name, ip, port, err := inf.CreateRegistion()
	if err != nil {
		return err
	}

	svcID := fmt.Sprintf("%s-%s-%d", name, ip, port)
	err = c.consul.Agent().ServiceDeregister(svcID)
	log.Infof("DeRegister service, param:%+v, error: %+v", svcID, err)

	return nil
}

// Service returns *model.Service
func (c *Client) Service(serviceName string) (*inf.Service, error) {
	return nil, fmt.Errorf("not supported")
}

// Instance returns *model.Instance
func (c *Client) Instance(serviceName string) (*inf.Instance, error) {
	entries, _, err := c.consul.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}
	if len(entries) < 1 {
		return nil, fmt.Errorf("not found")
	}

	instance := entries[0].Service
	return &inf.Instance{
		IP:       instance.Address,
		Port:     uint64(instance.Port),
		Metadata: instance.Meta,
	}, nil
}

// Instances returns []model.Instance
func (c *Client) Instances(serviceName string, filter map[string]interface{}) ([]inf.Instance, error) {
	entries, _, err := c.consul.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}

	result := make([]inf.Instance, 0, 10)
	for _, entry := range entries {
		result = append(result, inf.Instance{
			IP:       entry.Service.Address,
			Port:     uint64(entry.Service.Port),
			Metadata: entry.Service.Meta,
		})
	}
	return result, nil
}

func (c *Client) InstanceAddress(serviceName string) string {
	inst, err := c.Instance(serviceName)
	if err != nil {
		log.Errorf("Failed to get instance address, %v", err)
		return ""
	}
	return common.NormalAddress(inst.IP, int(inst.Port))
}

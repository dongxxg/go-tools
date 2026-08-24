package eureka

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dongxxg/go-tools/common"
	"github.com/dongxxg/go-tools/config"
	"github.com/dongxxg/go-tools/log"
	inf "github.com/dongxxg/go-tools/sd/interface"
	sd "github.com/dongxxg/go-tools/sd/interface"
)

// TODO 未测试
type AppInstance struct {
	InstanceID         string     `json:"instanceId"`
	HostName           string     `json:"hostName"`
	Port               Port       `json:"port"`
	App                string     `json:"app"`
	IPAddr             string     `json:"ipAddr"`
	Status             string     `json:"status"`
	SecurePort         SecurePort `json:"securePort"`
	LastDirtyTimestamp string     `json:"lastDirtyTimestamp"`
}

// pare port
//
//	eg."port": {
//					"$": 8080,
//					"@enabled": "true"
//				}
type Port struct {
	Port int `json:"$"`
}

// pare SecurePort
//
//	eg."securePort": {
//					"$": 8443,
//					"@enabled": "true"
//				}
type SecurePort struct {
	SecurePort int `json:"$"`
}

type Application struct {
	Name      string        `json:"name"`
	Instances []AppInstance `json:"instance"`
}

// Response for /eureka/apps/{appName}
type ServiceResponse struct {
	Application Application `json:"application"`
}

type Client struct {
	ServerAddr  string
	ServerPort  uint64
	Address     string
	Port        uint64
	ServiceName string
	InstanceID  string

	isRegistered bool
}

func NewClient() inf.Client {
	appName, localip, port, err := inf.CreateRegistion()
	if err != nil {
		return nil
	}

	return &Client{
		ServerAddr:  config.GetString("SD.Server.IpAddr"),
		ServerPort:  config.GetUint64("SD.Server.Port"),
		Address:     localip,
		Port:        port,
		ServiceName: appName,
		InstanceID:  fmt.Sprintf("%s:%s:%d", appName, localip, port),
	}
}

// Register 自动向Eureka注册服务，如果配置文件中有IP、Port配置，则使用配置文件中数据
func (c *Client) Register() error {
	cfg := c.newConfig(c.ServiceName, c.Address, c.Port)

	// define Register request
	registerAction := &common.RequestAction{
		URL:         fmt.Sprintf("http://%s/eureka/apps/%s", common.NormalAddress(c.ServerAddr, int(c.ServerPort)), c.ServiceName),
		Method:      "POST",
		ContentType: "application/json;charset=UTF-8",
		Accept:      "application/json;charset=UTF-8",
		Body:        cfg,
	}

	code, _, err := common.DoHTTP(registerAction)
	if err != nil || code != 204 {
		log.Errorf("register to eureka failed! param:%s, code:%v, error: %v", cfg, code, err)
		return fmt.Errorf("failed to register to eureka")
	}

	c.isRegistered = true
	go c.sendHeartBeat()

	log.Infoln("register to eureka successfully")
	return nil
}

func (c *Client) Deregister() error {
	c.isRegistered = false
	deregisterAction := &common.RequestAction{
		URL:    fmt.Sprintf("http://%s/eureka/apps/%s/%s", common.NormalAddress(c.ServerAddr, int(c.ServerPort)), c.ServiceName, c.InstanceID),
		Accept: "application/json;charset=UTF-8",
		Method: "DELETE",
	}

	code, rsp, err := common.DoHTTP(deregisterAction)
	if err != nil || code != 200 {
		resp := ""
		if rsp != nil {
			resp = string(rsp)
		}
		log.Errorf("deregister from eureka failed! instanceId:%s, code:%d, error: %v, response: %s", c.InstanceID, code, err, resp)
		return fmt.Errorf("failed to deregister from eureka")
	}

	log.Infof("deRegister service successfully, instanceId:%s", c.InstanceID)

	return nil
}

// Service returns *inf.Service
func (c *Client) Service(serviceName string) (*inf.Service, error) {
	return nil, fmt.Errorf("not supported")
}

// Instance returns *inf.Instance
func (c *Client) Instance(serviceName string) (*inf.Instance, error) {
	insts, err := c.Instances(serviceName, map[string]interface{}{
		"HealthyOnly": true,
	})
	if err != nil {
		return nil, err
	}
	if len(insts) == 0 {
		return nil, fmt.Errorf("can not find a healthy instance of service: %s", serviceName)
	}

	return &insts[0], nil
}

// Instances returns []inf.Instance
func (c *Client) Instances(serviceName string, filter map[string]interface{}) ([]inf.Instance, error) {
	// define get instance request
	requestAction := &common.RequestAction{
		URL:         fmt.Sprintf("http://%s/eureka/apps/%s", common.NormalAddress(c.ServerAddr, int(c.ServerPort)), serviceName),
		Method:      "GET",
		Accept:      "application/json;charset=UTF-8",
		ContentType: "application/json;charset=UTF-8",
	}
	log.Tracef("Query Eureka server using URL: %s", requestAction.URL)

	_, bytes, err := common.DoHTTP(requestAction)
	if err != nil {
		return nil, err
	} else if len(bytes) == 0 {
		return nil, fmt.Errorf("eureka returns empty result for service: %s", serviceName)
	}

	var m ServiceResponse
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		log.Errorf("parse JSON error(%v) from Eureka Server Response", err.Error())
		return nil, err
	}

	var healthOnly bool
	f, ok := filter["HealthyOnly"]
	if ok && f.(bool) {
		healthOnly = true
	}

	result := make([]inf.Instance, 0, 10)
	for _, inst := range m.Application.Instances {
		if healthOnly && !strings.EqualFold(inst.Status, "up") {
			continue
		}
		result = append(result, sd.Instance{
			IP:   inst.IPAddr,
			Port: uint64(inst.Port.Port),
		})
	}
	return result, nil
}

func (c *Client) InstanceAddress(serviceName string) string {
	inst, err := c.Instance(serviceName)
	if err != nil {
		log.Errorf("failed to get instance address, %v", err)
		return ""
	}
	return common.NormalAddress(inst.IP, int(inst.Port))
}

var configStr = `{
  "instance": {
	"instanceId" : "${appName}:${ipAddress}:${port}",
    "hostName":"${ipAddress}",
    "app":"${appName}",
    "ipAddr":"${ipAddress}",
    "vipAddress":"${appName}",
	"overriddenstatus": "UNKNOWN",
    "status":"UP",
    "port": {
      "$":${port},
      "@enabled": true
    },
	"securePort": {
      "$":443,
      "@enabled": false
    },
    "homePageUrl" : "http://${ipAddress}:${port}/",
    "statusPageUrl": "http://${ipAddress}:${port}/status",
    "healthCheckUrl": "http://${ipAddress}:${port}/health",
    "dataCenterInfo" : {
      "@class":"com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
      "name": "MyOwn"
    },
    "metadata": {
      "management.port" : "${port}"
    }
  }
}`

// newConfig load cfg from configStr
func (c *Client) newConfig(appName, localip string, port uint64) string {
	if localip == "" {
		localip = common.InitLocalIPbyUDP()
	}
	// load config
	cfg := configStr
	cfg = strings.Replace(cfg, "${ipAddress}", localip, -1)
	cfg = strings.Replace(cfg, "${port}", fmt.Sprint(port), -1)
	cfg = strings.Replace(cfg, "${appName}", appName, -1)
	return cfg
}

func (c *Client) sendHeartBeat() {
	registerAction := &common.RequestAction{
		URL:         fmt.Sprintf("http://%s/eureka/apps/%s/%s", common.NormalAddress(c.ServerAddr, int(c.ServerPort)), c.ServiceName, c.InstanceID),
		Method:      "PUT",
		ContentType: "application/json;charset=UTF-8",
	}

	// loop send heart beat every 5s
	for {
		if !c.isRegistered {
			break
		}
		code, _, err := common.DoHTTP(registerAction)
		if err != nil || code != 200 {
			log.Errorf("heartbeat to eureka failed! instanceId:%s, code:%d, error: %v", c.InstanceID, code, err)
		}
		time.Sleep(time.Second * 5)
	}
	log.Infoln("eureka heartbeat stopped")
}
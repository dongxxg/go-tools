package nacos

import (
	"net"

	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/dongxxg/go-tools/common"
	"github.com/dongxxg/go-tools/config"
	"github.com/dongxxg/go-tools/log"
	inf "github.com/dongxxg/go-tools/sd/interface"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"strings"
)

type Client struct {
	NamingClient naming_client.INamingClient
}

func NewClient() inf.Client {
	c, e := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": newServerConfig(),
		"clientConfig":  newClientConfig(),
	})
	if e != nil {
		log.Errorf("Failed to create Nacos client: %s", e.Error())
		return nil
	}

	return &Client{
		NamingClient: c,
	}
}

// Register 自动向Nacos注册服务，如果配置文件中有IP、Port配置，则使用配置文件中数据
func (c *Client) Register() error {
	name, ip, port, err := inf.CreateRegistion()
	if err != nil {
		return err
	}

	// 该字段表示注册的实例是否是临时实例还是持久化实例。如果是临时实例，则不会在 Nacos 服务端持久化存储，
	// 需要通过上报心跳的方式进行包活，如果一段时间内没有上报心跳，则会被 Nacos 服务端摘除。
	// 在被摘除后如果又开始上报心跳，则会重新将这个实例注册。持久化实例则会持久化被 Nacos 服务端，
	// 此时即使注册实例的客户端进程不在，这个实例也不会从服务端删除，只会将健康状态设为不健康。
	isEphemeral := config.GetBool("SD.Client.HeartBeat")
	param := vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: name,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   isEphemeral,
		// ClusterName: "DEFAULT",
		// GroupName:   "DEFAULT_GROUP",
		//Metadata: map[string]string{"idc": "shanghai"},
	}
	metadata := make(map[string]string)
	metadata["application.name"] = name                                                        //程序启动名称
	metadata["application.startup-time"] = common.GetStartTime().Format("2006-01-02 15:04:05") //程序启动时间
	releaseVersiondata, err := common.ReadReleaseVersion()
	if err == nil {
		metadata["pipeline.name"] = releaseVersiondata["pipelineName"]
		metadata["pipeline.app.version"] = releaseVersiondata["appVersion"]
		metadata["pipeline.app.build-date"] = releaseVersiondata["buildDate"]
		metadata["pipeline.project.name"] = releaseVersiondata["projectName"]
		metadata["pipeline.product.name"] = releaseVersiondata["productName"]
		metadata["pipeline.module.name"] = releaseVersiondata["moduleName"]
		metadata["pipeline.branch.name"] = releaseVersiondata["branchName"] //branchName
	}
	param.Metadata = metadata
	success, err := c.NamingClient.RegisterInstance(param)
	if err != nil || !success {
		log.Errorf("Register service failed! param:%+v, result:%v, error: %s", param, success, err.Error())
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

	isEphemeral := config.GetBool("SD.Client.HeartBeat")
	param := vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: name,
		Ephemeral:   isEphemeral,
		// Cluster:     "DEFAULT",
		// GroupName:   "DEFAULT_GROUP",
	}

	success, err := c.NamingClient.DeregisterInstance(param)
	if err != nil {
		log.Errorf("de-register service, param:%+v, result:%v, error: %s", param, success, err.Error())
	}
	log.Infof("de-register service successfully, param:%+v", param)

	return err
}

// Service returns *model.Service
func (c *Client) Service(serviceName string) (*inf.Service, error) {
	param := vo.GetServiceParam{
		ServiceName: serviceName,
		Clusters:    []string{"DEFAULT"},
		GroupName:   "DEFAULT_GROUP",
	}

	service, err := c.NamingClient.GetService(param)
	if err != nil {
		log.Errorf("sd Get Service error: %s", err.Error())
		return nil, err
	}

	return &inf.Service{
		Name:     service.Name,
		Metadata: service.Metadata,
	}, nil
}

// Instance returns *model.Instance
func (c *Client) Instance(serviceName string) (*inf.Instance, error) {
	param := vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
	}
	// SelectOneHealthyInstance return one instance by WRR strategy for load balance
	// And the instance should be health=true,enable=true and weight>0
	instance, err := c.NamingClient.SelectOneHealthyInstance(param)
	if err != nil {
		log.Errorf("Get %s instance error: %s", serviceName, err.Error())
		return nil, err
	}
	return &inf.Instance{
		IP:       instance.Ip,
		Port:     instance.Port,
		Metadata: instance.Metadata,
	}, nil
}

// Instances returns []model.Instance
func (c *Client) Instances(serviceName string, filter map[string]interface{}) ([]inf.Instance, error) {
	var instances []model.Instance
	var err error

	if healthOnly, ok := filter["HealthyOnly"]; ok {
		if healthOnly.(bool) {
			// SelectInstances only return the instances of healthy=${HealthyOnly},enable=true and weight>0
			param := vo.SelectInstancesParam{
				ServiceName: serviceName,
				HealthyOnly: true, // options.HealthOnly,
			}
			instances, err = c.NamingClient.SelectInstances(param)
			if err != nil {
				log.Errorf("Get %s instances error: %s", serviceName, err.Error())
				return nil, err
			}
		}
	} else {
		// SelectAllInstance return all instances,include healthy=false,enable=false,weight<=0
		param := vo.SelectAllInstancesParam{
			ServiceName: serviceName,
		}
		instances, err = c.NamingClient.SelectAllInstances(param)
		if err != nil {
			log.Errorf("Get %s instances error: %s", serviceName, err.Error())
			return nil, err
		}
	}

	result := make([]inf.Instance, 0, 10)
	for _, inst := range instances {
		result = append(result, inf.Instance{
			IP:       inst.Ip,
			Port:     inst.Port,
			Metadata: inst.Metadata,
		})
	}

	return result, nil
}

func (c *Client) InstanceAddress(serviceName string) string {
	inst, err := c.Instance(serviceName)
	if err != nil {
		log.Errorf("Failed to get instance address, %s", err.Error())
		return ""
	}
	//bug 162924 适配v6地址带[]的ip
	if strings.HasPrefix(inst.IP, "[") && strings.HasSuffix(inst.IP, "]") {
		inst.IP = strings.TrimLeft(inst.IP, "[")
		inst.IP = strings.TrimRight(inst.IP, "]")
	}
	return common.NormalAddress(inst.IP, int(inst.Port))
}

func newClientConfig() constant.ClientConfig {
	logDir := config.GetString("Logging.Path")
	log.Infof("Nacos will be logged to directory: %s", logDir)

	return constant.ClientConfig{
		// NamespaceId:         config.GetString("SD.Client.NamespaceId"),
		TimeoutMs:            config.GetUint64("SD.Client.TimeoutMs"),
		NotLoadCacheAtStart:  true,
		UpdateCacheWhenEmpty: true,
		LogDir:               logDir,
		LogLevel:             "warn",
		CacheDir:             "./nacos_cache",
		Username:             config.GetString("SD.Client.Username"),
		Password:             config.GetString("SD.Client.Password"),
	}
}

func newServerConfig() []constant.ServerConfig {
	serverIp := config.GetString("SD.Server.IpAddr")
	addr := net.ParseIP(serverIp)
	// is IPv6
	if addr != nil && addr.To4() == nil {
		serverIp = "[" + serverIp + "]"
	}
	return []constant.ServerConfig{
		// 可以配置多个ServerConfig，进行负载分担，目前只配置了一个
		{
			IpAddr: serverIp,
			// ContextPath: config.GetString("SD.Server.ContextPath"),
			Port: config.GetUint64("SD.Server.Port"),
		},
	}
}
package sd

type Client interface {
	// 服务注册
	Register() error

	// 服务反注册
	Deregister() error

	// 返回服务信息
	Service(serviceName string) (*Service, error)

	// 返回一个健康实例
	Instance(serviceName string) (*Instance, error)

	// 返回一个健康实例的IP:端口
	InstanceAddress(serviceName string) string

	// 根据过滤条件返回实例
	Instances(serviceName string, filter map[string]interface{}) ([]Instance, error)
}

type Instance struct {
	Port     uint64            `json:"port"`
	IP       string            `json:"ip"`
	Metadata map[string]string `json:"metadata"`
}

type Service struct {
	Metadata map[string]string `json:"metadata"`
	Name     string            `json:"name"`
}

//go:build !iot

package config

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"net"
)

type nacosConfigClient struct {
	client config_client.IConfigClient
}

// GetConfig is Nacos GetConfig proxy
func (s *nacosConfigClient) GetConfig(grp, key string) (string, error) {
	return s.client.GetConfig(vo.ConfigParam{
		DataId: key,
		Group:  grp,
	})
}

// PublishConfig is Nacos GetConfig proxy
func (s *nacosConfigClient) PublishConfig(grp, key, content string) (bool, error) {
	return s.client.PublishConfig(vo.ConfigParam{
		DataId:  key,
		Group:   grp,
		Content: content,
	})
}

// ListenConfig wraps Nacos ListenConfig
func (s *nacosConfigClient) ListenConfig(grp, key string, cb func(namespace, group, dataId, data string)) error {
	return s.client.ListenConfig(vo.ConfigParam{
		DataId:   key,
		Group:    grp,
		OnChange: cb,
	})
}

// CancelListenConfig wraps Nacos ListenConfig
func (s *nacosConfigClient) CancelListenConfig(grp, key string) error {
	return s.client.CancelListenConfig(vo.ConfigParam{
		DataId: key,
		Group:  grp,
	})
}
func initNacosConfigClient() (IConfigCenter, error) {
	serverIP := GetString("SD.Server.IpAddr")
	addr := net.ParseIP(serverIP)
	// is IPv6
	if addr != nil && addr.To4() == nil {
		serverIP = "[" + serverIP + "]"
	}
	sc := []constant.ServerConfig{
		{
			IpAddr: serverIP,
			Port:   GetUint64("SD.Server.Port"),
		},
	}

	logDir := GetString("Logging.Path")
	cc := constant.ClientConfig{
		//NamespaceId:         "e525eafa-f7d7-4029-83d9-008937f9d468", //namespace id
		TimeoutMs:            GetUint64("SD.Client.TimeoutMs"),
		NotLoadCacheAtStart:  true,
		UpdateCacheWhenEmpty: true,
		LogDir:               logDir,
		LogLevel:             "warn",
		CacheDir:             "./nacos_cache",
		Username:             GetString("SD.Client.Username"),
		Password:             GetString("SD.Client.Password"),
	}

	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc,
		"clientConfig":  cc,
	})

	if err != nil {
		return nil, err
	}

	fmt.Println("uni-base ==> Nacos config client initialized!")

	return &nacosConfigClient{
		client: client,
	}, nil
}

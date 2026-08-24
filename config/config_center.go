//go:build !iot

package config

import (
	"errors"
	"fmt"
)

type IConfigCenter interface {
	GetConfig(grp, key string) (string, error)
	PublishConfig(grp, key, content string) (bool, error)
	ListenConfig(grp, key string, cb func(namespace, group, dataId, data string)) error
	CancelListenConfig(grp, key string) error
}

var configClient IConfigCenter

func GetConfigClient() (IConfigCenter, error) {
	if configClient != nil {
		return configClient, nil
	}
	var err error
	sdType := GetStringWithDefault("SD.Type", "nacos")
	if sdType == "nacos" {
		configClient, err = initNacosConfigClient()
	} else if sdType == "consul" {
		configClient, err = initConsulConfigClient()
	} else {
		return nil, errors.New("invalid config type: " + sdType)
	}
	return configClient, err
}

// GetConfig is Nacos GetConfig proxy
func GetConfig(grp, key string) (string, error) {
	if configClient == nil {
		return "", fmt.Errorf("config center not exist")
	}

	return configClient.GetConfig(grp, key)
}

// PublishConfig is Nacos GetConfig proxy
func PublishConfig(grp, key, content string) (bool, error) {
	if configClient == nil {
		return false, fmt.Errorf("config center not exist")
	}

	return configClient.PublishConfig(grp, key, content)
}

// ListenConfig wraps Nacos ListenConfig
func ListenConfig(grp, key string, cb func(namespace, group, dataId, data string)) error {
	if configClient == nil {
		return fmt.Errorf("config center not exist")
	}

	return configClient.ListenConfig(grp, key, cb)
}

// CancelListenConfig wraps Nacos ListenConfig
func CancelListenConfig(grp, key string) error {
	if configClient == nil {
		return fmt.Errorf("config center not exist")
	}

	return configClient.CancelListenConfig(grp, key)
}

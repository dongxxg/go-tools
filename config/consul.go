//go:build !iot

package config

import (
	"errors"
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"net"
	"strconv"
	"sync"
	"time"
)

type consulConfigClient struct {
	client *consulapi.Client
}

func getKey(grp, key string) string {
	if grp != "" {
		return grp + "/" + key
	}
	return key
}
func (c *consulConfigClient) GetConfig(grp, key string) (string, error) {
	kv := c.client.KV()
	pair, _, err := kv.Get(getKey(grp, key), nil)
	if err != nil || pair == nil {
		return "", err
	}
	return string(pair.Value), err
}

func (c *consulConfigClient) PublishConfig(grp, key, content string) (bool, error) {
	kv := c.client.KV()
	_, err := kv.Put(&consulapi.KVPair{
		Key:   getKey(grp, key),
		Value: []byte(content),
	}, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

var listenConfigMap = sync.Map{}

func (c *consulConfigClient) ListenConfig(grp, key string, cb func(namespace string, group string, dataId string, data string)) error {
	closeChan := make(chan struct{}) // fixme 通过watch监听还得获取自己的ip和端口，还得写接口获取返回值，过于复杂。
	_, found := listenConfigMap.Load(getKey(grp, key))
	if found {
		return errors.New("listen already existed")
	}
	listenConfigMap.Store(getKey(grp, key), closeChan)

	watchParams := &consulapi.QueryOptions{WaitIndex: 0}
	go func() {
		for {
			select {
			case <-closeChan:
				return
			case <-time.After(time.Second * 30):
				kv := c.client.KV()
				// 使用Watch方法监听键的变化
				pair, meta, err := kv.Get(key, watchParams)
				if err != nil {
					fmt.Println("consul listen config error:", err.Error())
					continue
				}
				if pair == nil {
					continue
				}
				// 检查配置是否发生变化
				if meta.LastIndex != watchParams.WaitIndex {
					fmt.Printf("Value for key '%s' changed to: %s\n", key, string(pair.Value))
					watchParams.WaitIndex = meta.LastIndex
					cb("", grp, "", string(pair.Value))
				}
			}
		}
	}()
	return nil
}

func (c *consulConfigClient) CancelListenConfig(grp, key string) error {
	ch, found := listenConfigMap.LoadAndDelete(getKey(grp, key))
	if !found {
		return errors.New("listen cancel  listenConfig not found")
	}
	close(ch.(chan struct{}))
	return nil
}

func initConsulConfigClient() (IConfigCenter, error) {
	config := consulapi.DefaultConfig()

	serverIP := GetString("SD.Server.IpAddr")
	addr := net.ParseIP(serverIP)
	// is IPv6
	if addr != nil && addr.To4() == nil {
		serverIP = "[" + serverIP + "]"
	}
	config.Address = serverIP + ":" + strconv.Itoa(int(GetUint64("SD.Server.Port")))
	username := GetString("SD.Client.Username")
	password := GetString("SD.Client.Password")
	if username != "" && password != "" {
		config.HttpAuth = &consulapi.HttpBasicAuth{
			Username: username,
			Password: password,
		}
	} else if password != "" {
		config.Token = GetString("SD.Client.Password")
	}
	client, err := consulapi.NewClient(config)
	if err == nil {
		fmt.Println("uni-base ==> Nacos config client initialized!")
	}
	cc := consulConfigClient{
		client: client,
	}
	return &cc, err
}

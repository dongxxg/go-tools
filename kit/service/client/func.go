package client

import (
	"github.com/pkg/errors"
	"unitechs.com/cno/kit/service/conf"
	"unitechs.com/unios-dice/uni-base/core/config"
)

// ListenNacos 监听nacos key 的变化
func ListenNacos(getter IConfigGetter, key string, callBack func(namespace, group, dataId, data string)) error {
	nacosKey := conf.GenNacosKey(key)

	if err := config.CancelListenConfig(getter.NacosGroup(), nacosKey); err != nil {
		return errors.Wrap(err, "CancelListenConfig ")
	}

	if err := config.ListenConfig(getter.NacosGroup(), nacosKey, callBack); err != nil {
		return errors.Wrapf(err, "ListenConfig %s", nacosKey)
	}
	return nil
}

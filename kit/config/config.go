package config

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"unitechs.com/unios-dice/uni-base/core/common"
	"unitechs.com/unios-dice/uni-base/core/config"
)

func Init(confPath, fileName string) {
	if confPath == "" {
		os.Getppid()
		binPath, err := os.Executable()
		if err != nil {
			panic("fatal error when getting current binary path")
		}
		//从环境变量读取
		filePath := os.Getenv("UNIBASE_CONFIG_PATH")
		if filePath == "" {
			//从命令行获取
			if common.WithinUT() { // 如果是测试则不从命令行读取配置
				return
			}
			// 根据工程目录规范，跟二进制并列有conf目录
			confDir := filepath.Dir(binPath)
			// 判断是否在debugging
			if strings.Contains(binPath, "__debug_bin") {
				confDir, _ = os.Getwd()
			}
			// 这样在从其它目录启动可执行文件时，能够找到conf目录
			confPath = filepath.Join(confDir, "conf")
			filePath = filepath.Join(confPath, fileName)
		} else {
			confPath = filepath.Dir(filePath)
			fileName = filepath.Base(filePath)
		}
		_, err = os.Stat(filePath)
		if err != nil && os.IsNotExist(err) {
			// 如果默认配置文件不存在，则必须在主函数中初始化配置才能使用
			// fmt.Println("default configuration file not found in: ", confPath)
			return
		}

		viper.Set("Service.BinaryFile", binPath)
	}

	config.Init(confPath, fileName)
}

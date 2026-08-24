package redis

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"time"
)

type Redis struct {
	Host     string `mapstructure:"addr" json:"addr" yaml:"addr"`             // 服务器地址:端口
	Password string `mapstructure:"password" json:"password" yaml:"password"` // 密码
	MaxIdle  int `mapstructure:"maxIdle" yaml:"maxIdle"`
	MaxActive int `mapstructure:"maxActive" yaml:"maxActive"`
	IdleTimeout time.Duration `mapstructure:"idleTimeout"`
}

type Server struct {
	Redis Redis
}

var CONF Server

func Viper() *viper.Viper{
	v := viper.New()
	v.SetConfigFile("./redis.yaml")
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		log.Fatalf("Fail to read yaml:%v", err)
	}

	v.WatchConfig()

	v.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("Config file changed:", in.Name)
		if err = v.Unmarshal(&CONF);err!= nil {
			fmt.Println(err)
		}
	})

	if err = v.Unmarshal(&CONF); err != nil {
		fmt.Println(err)
	}
	return v
}


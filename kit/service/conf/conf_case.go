package conf

import (
	"fmt"
	"github.com/spf13/viper"
)

func NewTestKitConf(path string) *KitConf {
	NewTestConfig(path)
	return &KitConf{
		Viper: viper.GetViper(),
	}
}

func NewTestConfig(path string) {
	viper.SetConfigName("configuration") // name of config file (without extension)
	viper.SetConfigType("toml")          // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(path)            // path to look for the config file in
	viper.AddConfigPath(".")             // optionally look for config in the working directory
	err := viper.ReadInConfig()          // Find and read the config file
	if err != nil {                      // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	a := viper.GetString("Event.Kafka.Brokers")
	fmt.Println(a)

}

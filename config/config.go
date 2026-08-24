//go:build !iot

// Package config 配置模块。注意，config包不要依赖dice的log！因为log依赖于config。
package config

import (
	"flag"
	"fmt"
	"github.com/dongxxg/go-tools/common"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var loadTime time.Time

// 如果默认配置文件存在，则直接初始化配置，应用程序可以通过调用Init()方法覆盖默认配置
func init() {
	os.Getppid()
	binPath, err := os.Executable()
	if err != nil {
		panic("fatal error when getting current binary path")
	}
	confPath := ""
	fileName := "configuration.toml"
	//从环境变量读取
	filePath := os.Getenv("CONFIG_PATH")
	if filePath == "" {
		//从命令行获取
		if common.WithinUT() { // 如果是测试则不从命令行读取配置
			return
		}
		filePath1 := flag.String("config", "", "Specify the configuration file or folder")
		flag.Parse()
		filePath = *filePath1
		if filePath == "" {
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
	Init(confPath, fileName)
}

// Init 初始化配置模块。如果根据工程目录规范的默认路径下有配置文件，在模块初始化时自动加载配置，
// 否则应用程序需要显示地初始化配置模块。初始化参数：path - 配置文件所在目录，name - 配置文件名（包括后缀）。
func Init(path, name string) {
	var fileName, ext string
	ext = filepath.Ext(name)
	fileName = name[0 : len(name)-len(ext)]
	if ext == "" {
		panic(fmt.Sprintf("config type could not be determined for %s", name))
	}

	viper.SetConfigName(fileName) // name of config file (without extension)
	viper.SetConfigType(ext[1:])  // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(path)
	// viper.AddConfigPath(".") // optionally look for config in the working directory

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(err)
	}

	viper.Set("Service.ConfigFile", viper.ConfigFileUsed())

	loadTime = time.Now()
	// 监视配置文件变化，会自动触发配置重新加载
	viper.WatchConfig()
	// 绑定环境变量
	bindEnv()
	bindRemoteConfig()
}

// InitConfigCenter reads config from config center
func InitConfigCenter() error {
	_, err := GetConfigClient()
	return err
}

// AddConfigPath adds a path for Viper to search for the config file in.
// Can be called multiple times to define multiple search paths.
func AddConfigPath(in string) {
	viper.AddConfigPath(in)
}

// SetConfigType sets the type of the configuration returned by the
// remote source, e.g. "json".
func SetConfigType(in string) {
	viper.SetConfigType(in)
}

// SetConfigName sets name for the config file.
// Does not include extension.
func SetConfigName(in string) {
	viper.SetConfigName(in)
}

// Set sets the value for the key in the override register.
// Set is case-insensitive for a key.
// Will be used instead of values obtained via
// flags, config file, ENV, default, or key/value store.
func Set(key string, value interface{}) { viper.Set(key, value) }

// SetDefault sets the default value for this key.
// SetDefault is case-insensitive for a key.
func SetDefault(key string, value interface{}) {
	viper.SetDefault(key, value)
}

// Get can retrieve any value given the key to use.
func Get(key string) interface{} {
	return viper.Get(key)
}

// GetWithDefault returns the value associated with the key, if key not exist, return default value.
func GetWithDefault(key string, d interface{}) interface{} {
	if viper.IsSet(key) {
		return viper.Get(key)
	}
	return d
}

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetInt64 returns the value associated with the key as an integer.
func GetInt64(key string) int64 {
	return viper.GetInt64(key)
}

// GetIntSlice returns the value associated with the key as a slice of int values.
func GetIntSlice(key string) []int {
	return viper.GetIntSlice(key)
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func GetUint64(key string) uint64 {
	return viper.GetUint64(key)
}

// GetBool returns the value associated with the key as a boolean.
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string {
	return viper.GetString(key)
}

func GetStringAndNacos(key string) string {
	var data string
	var err error
	//环境变量先拿
	keyenv := strings.Replace(strings.ToUpper(key), ".", "_", -1)
	data = os.Getenv(keyenv)
	if data != "" {
		return data
	}
	//然后去nacos
	data, err = GetConfig("PUB", keyenv)
	if data != "" && err == nil {
		return data
	}
	return viper.GetString(key)
}

// GetStringWithDefault returns the value associated with the key as a string, if key not exist, return default value.
func GetStringWithDefault(key, d string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	return d
}

// GetStringSlice returns the value associated with the key as a slice of strings.
// 由于k8s不支持设置数组型环境变量，慎用此方法，建议用逗号分隔字符串来代替
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

// GetTime returns the value associated with the key as time.
func GetTime(key string) time.Time {
	return viper.GetTime(key)
}

// GetStringMapString returns the value associated with the key as a map of strings.
func GetStringMapString(key string) map[string]string {
	return viper.GetStringMapString(key)
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func GetStringMapStringSlice(key string) map[string][]string {
	return viper.GetStringMapStringSlice(key)
}

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}

// BindEnv binds a key to a ENV variable.
// ENV variables are case sensitive.
func BindEnv(input ...string) error {
	return viper.BindEnv(input...)
}

// IsSet checks to see if the key has been set in any of the data locations.
func IsSet(key string) bool { return viper.IsSet(key) }

// WriteConfigAs writes current configuration to a given filename.
func WriteConfigAs(filename string) error {
	return viper.WriteConfigAs(filename)
}

// AllSettings merges all settings and returns them as a map[string]interface{}.
func AllSettings() map[string]interface{} {
	return viper.AllSettings()
}

// WatchConfig tells viper to watch configuration changes
func WatchConfig() { viper.WatchConfig() }

// OnConfigChange will call the input func when configuration changes
func OnConfigChange(run func(in fsnotify.Event)) { viper.OnConfigChange(run) }

// UnmarshalKey takes a single key and marshals it into a Struct
func UnmarshalKey(key string, rawVal interface{}) error {
	return viper.UnmarshalKey(key, rawVal)
}

// GetConfFile returns the file used to populate the config registry.
func GetConfFile() string {
	return viper.ConfigFileUsed()
}

func GetLoadTime() time.Time {
	return loadTime
}

func SetLoadTime() {
	loadTime = time.Now()
}

func bindEnv() {
	keys := viper.AllKeys()
	for _, key := range keys {
		envVar := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		viper.BindEnv(key, envVar)
	}
}

// 从nacos读取所有配置，如果有值则覆盖本地配置
func bindRemoteConfig() {
	waitGroup := sync.WaitGroup{}
	m := sync.Map{}
	group := viper.GetString("SD.Client.Group")
	err := InitConfigCenter()
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	keys := viper.AllKeys()
	for index := range keys {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			envVar := strings.ToUpper(strings.ReplaceAll(keys[index], ".", "_"))
			if os.Getenv(envVar) != "" {
				return
			}
			config, err := GetConfig(group, envVar)
			if err != nil || len(config) == 0 {
				return
			}
			m.Store(keys[index], config)
		}(index)
	}
	waitGroup.Wait()
	m.Range(func(key, value interface{}) bool {
		viper.Set(key.(string), value.(string))
		return true
	})
	return
}
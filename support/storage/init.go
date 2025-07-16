/**
 * @author  zhaoliang.liang
 * @date  2025/1/22 15:26
 */

package storage

import (
	"gorm.io/gorm"
	"support/logger"
	"support/storage/config"
	"support/storage/driver"
	"support/storage/manager"
)

// UrlPaths 不同对象存储驱动对应的Url信息
var UrlPaths []*config.URLPath

// GlobalStorageConfigs 全局对象存储配置
var GlobalStorageConfigs *config.OssConfigs

// OssHelper 对象存储助手
var OssHelper *driver.StorageHelper

// initStorageConfig 初始化对象存储全局配置
func initStorageConfig(configs []*config.StorageConfig) {
	if len(configs) == 0 {
		panic("生成ossConfig失败,storageConfig为空")
	}
	var ossConfig []*config.StorageConfig
	for _, preConf := range configs {
		ossConfig = append(ossConfig, preConf)
	}
	GlobalStorageConfigs = &config.OssConfigs{
		Configs: ossConfig,
	}

}

// GetGetCloudyUrlPathFunc 获取GetCloudyUrlPathFunc
func GetGetCloudyUrlPathFunc(db *gorm.DB) manager.GetCloudyUrlPath {
	getDb := func() *gorm.DB {
		return db
	}
	return manager.GetGetCloudyUrlPathFunc(getDb)
}

// InitOssManual 手动初始化对象存储实例
func InitOssManual(storageConfig *config.StorageConfig) {
	var err error
	getStorageConfigFunc := func() *config.StorageConfig {
		storageConfig := storageConfig
		if storageConfig.Internal && storageConfig.EndpointInternal != "" {
			storageConfig.Endpoint = storageConfig.EndpointInternal
		}
		return storageConfig
	}

	OssHelper, err = manager.InitSpecialManager(getStorageConfigFunc)
	if err != nil {
		logger.Error("InitSpecialOss InitOssManual error %s ", err)
		panic(err)
	}
}

// InitOssByInsert 通过插入初始化对象存储实例
func InitOssByInsert(helper *driver.StorageHelper) {
	OssHelper = helper
}

// InitStorageManagerManual 手动初始化对象管理器
func InitStorageManagerManual(urlPaths []*config.URLPath, configs config.OssConfigs) {
	getCloudyUrlPath := func() []*config.URLPath {
		return urlPaths
	}
	err := manager.InitCloudyManager(getCloudyUrlPath, manager.GetStorageConfigs(&configs))
	if err != nil {
		logger.Error("InitSpecialOss InitCloudyManual error %s ", err)
		panic(err)
	}
}

// InitStorageManagerFromInternal 从school_internal获取配置并初始化
func InitStorageManagerFromInternal() error {
	logger.Error("please use InitStorageManagerFromDb to init oss")
	// 多云配置
	getUrlPath, getConfigs, err := manager.GetStoragePreBySchoolInternal()
	if err != nil {
		return err
	}
	UrlPaths = getUrlPath()
	initStorageConfig(getConfigs())
	// 多云使用
	err = manager.InitCloudyManager(
		getUrlPath,
		getConfigs,
	)
	if err != nil {
		return err
	}
	return nil
}

// InitStorageManagerFromDb 从db获取配置并初始化
func InitStorageManagerFromDb(db *gorm.DB, configs config.OssConfigs) error {
	err := manager.InitCloudyManager(
		manager.GetGetCloudyUrlPathFunc(func() *gorm.DB {
			return db
		}),
		manager.GetStorageConfigs(&configs),
	)
	if err != nil {
		return err
	}
	return nil
}

// InitStorageHelper 初始化对象存储助手
func InitStorageHelper(storageConfig *config.StorageConfig) (*driver.StorageHelper, error) {
	getStorageConfigFunc := func() *config.StorageConfig {
		return storageConfig
	}
	return manager.InitSpecialManager(getStorageConfigFunc)
}

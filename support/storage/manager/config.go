package manager

import (
	pb "api/storage"
	"context"
	"ginmw/httpRpc/rpcService"
	"gorm.io/gorm"
	"support"
	"support/storage/config"
	"tableModel"
)

type GetDbFunc func() *gorm.DB

func GetStorageConfigs(inputConfigs *config.OssConfigs) GetCloudyStorageConfigs {
	// 初始化通用的对象存储助手
	// 构建不同云的配置
	var configs []*config.StorageConfig
	for _, preConf := range inputConfigs.Configs {
		configs = append(configs, preConf)
	}
	return func() []*config.StorageConfig {
		return configs
	}
}

func GetGetCloudyUrlPathFunc(getDb GetDbFunc) GetCloudyUrlPath {
	db := getDb()
	if db == nil {
		return nil
	}
	list := getAllUrlPathInfo(db)
	var res []*config.URLPath
	for _, item := range list {
		res = append(res, &config.URLPath{
			ID:         item.ID,
			Name:       item.Name,
			UniqueName: item.UniqueName,
			Region:     item.Region,
			Bucket:     item.Bucket,
			Path:       item.Path,
			Provider:   item.Provider,
			Host:       item.Host,
			Note:       item.Note,
		})
	}
	return func() []*config.URLPath {
		return res
	}
}

func getAllUrlPathInfo(db *gorm.DB) []*tableModel.URLPath {
	var res []*tableModel.URLPath
	dbIns := db.Model(&tableModel.URLPath{}).
		Where("is_active=1").
		Find(&res)
	if dbIns.Error != nil {
		return nil
	}
	return res
}

func GetStoragePreBySchoolInternal() (getUrlPaths GetCloudyUrlPath, getStorageConfigs GetCloudyStorageConfigs, err error) {
	ctx := context.Background()
	resp, err := rpcService.GetStorageUrlPathConfig(ctx, &pb.EmptyRequest{})
	if err != nil {
		return nil, nil, err
	}
	getUrlPaths = GetGetCloudyUrlPathFuncByPb(resp.UrlPaths)
	getStorageConfigs = GetStorageConfigsByPb(resp.StorageConfigs)
	return
}

func GetStorageConfigsByPb(inputConfigs []*pb.StorageConfig) GetCloudyStorageConfigs {
	// 初始化通用的对象存储助手
	// 构建不同云的配置
	var configs []*config.StorageConfig
	for _, preConf := range inputConfigs {
		storageConfig := &config.StorageConfig{
			Provider:         preConf.Provider,
			Protocol:         preConf.Protocol,
			AccessKeyId:      preConf.AccessKeyId,
			AccessKeySecret:  preConf.AccessKeySecret,
			Root:             preConf.Root,
			BaseDir:          "",
			RoleArn:          preConf.RoleArn,
			Bucket:           preConf.Bucket,
			Region:           preConf.Region,
			Internal:         support.BoolStr(preConf.Internal),
			Endpoint:         preConf.Endpoint,
			EndpointInternal: preConf.EndpointInternal,
			StsEndPoint:      preConf.StsEndpoint,
			Host:             preConf.Host,
			CustomHost:       preConf.CustomHost,
			Path:             preConf.Path,
			TmpPath:          preConf.TmpPath,
			TmpRoot:          preConf.TmpRoot,
			CdnDomain:        preConf.CdnDomain,
			CdnProtocol:      preConf.CdnProtocol,
		}
		configs = append(configs, storageConfig)
	}
	return func() []*config.StorageConfig {
		return configs
	}
}

func GetGetCloudyUrlPathFuncByPb(ups []*pb.UrlPath) GetCloudyUrlPath {
	var res []*config.URLPath
	for _, item := range ups {
		res = append(res, &config.URLPath{
			ID:         item.Id,
			Name:       item.Name,
			UniqueName: item.UniqueName,
			Region:     item.Region,
			Bucket:     item.Bucket,
			Path:       item.Path,
			Provider:   item.Provider,
			Host:       &item.Host,
			Note:       &item.Note,
		})
	}
	return func() []*config.URLPath {
		return res
	}
}

func GetPbUrlPaths(db *gorm.DB) []*pb.UrlPath {
	list := getAllUrlPathInfo(db)
	var res []*pb.UrlPath
	for _, urlPath := range list {
		u := &pb.UrlPath{}
		u.Id = urlPath.ID
		u.UniqueName = urlPath.UniqueName
		u.Name = urlPath.Name
		u.Region = urlPath.Region
		u.Bucket = urlPath.Bucket
		u.Path = urlPath.Path
		u.Provider = urlPath.Provider
		if urlPath.Host != nil {
			u.Host = *urlPath.Host
		}
		if urlPath.Note != nil {
			u.Note = *urlPath.Note
		}
		res = append(res, u)
	}
	return res
}

func GetPbConfigs(inputConfigs *config.OssConfigs) []*pb.StorageConfig {
	var res []*pb.StorageConfig
	for _, conf := range inputConfigs.Configs {
		internal := conf.Internal.GetValue()
		c := &pb.StorageConfig{}
		c.Provider = conf.Provider
		c.Protocol = conf.Protocol
		c.AccessKeyId = conf.AccessKeyId
		c.AccessKeySecret = conf.AccessKeySecret
		c.Endpoint = conf.Endpoint
		c.EndpointInternal = conf.EndpointInternal
		c.StsEndpoint = conf.StsEndPoint
		c.Bucket = conf.Bucket
		c.RoleArn = conf.RoleArn
		c.Region = conf.Region
		c.Root = conf.Root
		c.TmpRoot = conf.TmpRoot
		c.Internal = internal
		c.Host = conf.Host
		c.CustomHost = conf.CustomHost
		c.CdnDomain = conf.CdnDomain
		c.CdnProtocol = conf.CdnProtocol
		c.Path = conf.Path
		c.TmpPath = conf.TmpPath
		res = append(res, c)
	}
	return res
}

package driver

import "time"

type Storage interface {
	GetObject(bucket string, key string) ([]byte, error)

	GetFile(bucket string, key string, localFile string) error

	PutObject(bucket string, key string, data []byte, metadata map[string]string) error

	PutObjectWithMeta(bucket string, key string, data []byte, metadata *Metadata) error

	PutFile(bucket string, key string, srcFile string, metadata map[string]string) error

	PutObjectFromFile(bucket string, key string, filePath string, metadata map[string]string) error

	PutFileWithMeta(bucket string, key string, srcFile string, metadata *Metadata) error

	// 大文件 断点续传  分片数量不能超过10000,分片大小推荐  100K-1G partSize单位KB
	PutFileWithPart(
		bucket string,
		key string,
		srcFile string,
		metadata *Metadata,
		partSize int64,
	) error

	ListObjects(bucket string, prefix string) ([]Content, error)

	DeleteObject(bucket string, key string) error

	CopyObject(bucket string, srcKey string, destKey string) error

	SetObjectAcl(bucket string, key string, acl StorageAcl) error

	SetObjectMetaData(bucket string, key string, metadata *Metadata) error

	IsObjectExist(bucket string, key string) (bool, error)

	//判断目录是否存在 true 存在 false 不存在或无子文件
	IsNotEmptyDirExist(bucket string, key string) (bool, error)

	IsPublicAcl(bucket string, key string) (bool, error)

	GetDirToken(remoteDir string) map[string]interface{}

	GetDirToken2(remoteDir string) *StorageToken

	GetDirTokenWithAction(remoteDir string, actions ...Action) (bool, *StorageToken)
	//过期时间：秒
	SignFile(remoteDir string, expiredTime int64) (error, string)

	//过期时间：秒
	SignFile2(bucket, remoteDir string, expiredTime int64) (error, string)

	//过期时间：秒
	SignFileForDownload(remoteDir string, expiredTime int64, downloadName string) string

	GetObjectMeta(bucket string, key string) (*Content, error)
	//解冻归档文件，成功就为true,异常或者失败返回false
	RestoreArchive(bucket string, key string) (bool, error)

	//判断是否归档文件
	IsArchive(bucket string, key string) (bool, error)

	//批量删除文件；可以减少调用次数，进而减少费用
	BatchDeleteObject(bucketName string, list []string) (successList []string, e error)
}

// listObject结果对象
type Content struct {
	Key          string
	Size         int64
	ETag         string
	LastModified time.Time
}

type StorageAcl string

const (
	AclPrivate         StorageAcl = "private"
	AclPublicRead      StorageAcl = "public-read"
	AclPublicReadWrite StorageAcl = "public-read-write" //2021.1.8当前测试华为public-read-write设置未生效
	AclDefault         StorageAcl = "default"           //华为不支持，请勿使用
)

type MimeType string

const (
	PDF_MimeType MimeType = "application/pdf"
)

type Metadata struct {
	Mime               string
	ContentEncoding    string
	Acl                string
	ContentDisposition string
}

type StorageToken struct {
	AccessKeyID     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	Bucket          string `json:"bucket"`
	Expire          int64  `json:"expire"`
	Host            string `json:"host"`
	EndPoint        string `json:"endPoint"`
	Port            int64  `json:"port"`
	Provider        string `json:"provider"`
	Region          string `json:"region"`
	StsToken        string `json:"stsToken"`
	UploadPath      string `json:"uploadPath"`
	Path            string `json:"path"`
	CdnDomain       string `json:"cdnDomain"`
}

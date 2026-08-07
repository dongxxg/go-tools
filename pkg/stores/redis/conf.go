package redis

import "errors"

var (
	// ErrEmptyHost is an error that indicates no redis host is set.
	ErrEmptyHost = errors.New("empty redis host")
	// ErrEmptyType is an error that indicates no redis type is set.
	ErrEmptyType = errors.New("empty redis type")
	// ErrEmptyKey is an error that indicates no redis key is set.
	ErrEmptyKey = errors.New("empty redis key")
)

type (
	// A RedisConf is a redis config.
	RedisConf struct {
		Host     string
		Type     string `json:",default=node,options=node|cluster"`
		Pass     string `json:",omitempty"`
		Tls      bool   `json:",default=false,options=true|false"`
		Db       int
		PoolSize int // 连接池大小
	}
)

// NewRedis returns a Redis.
func (rc RedisConf) NewRedis() *Redis {
	var opts []Option
	if rc.Type == ClusterType {
		opts = append(opts, Cluster())
	}
	if len(rc.Pass) > 0 {
		opts = append(opts, WithPass(rc.Pass))
	}
	if rc.Tls {
		opts = append(opts, WithTLS())
	}
	if rc.PoolSize != 0 {
		opts = append(opts, WithPoolSize(rc.PoolSize))
	}

	opts = append(opts, WithDb(rc.Db))

	return New(rc.Host, opts...)
}

// Validate validates the RedisConf.
func (rc RedisConf) Validate() error {
	if len(rc.Host) == 0 {
		return ErrEmptyHost
	}

	if len(rc.Type) == 0 {
		return ErrEmptyType
	}

	return nil
}

package redis

import (
	"crypto/tls"
	"io"
	"strings"

	red "github.com/go-redis/redis/v8"
	"unitechs.com/unios-dice/uni-base/lib/syncx"
)

var clusterManager = syncx.NewResourceManager()

func getCluster(r *Redis) (*red.ClusterClient, error) {
	val, err := clusterManager.GetResource(r.Addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		store := red.NewClusterClient(&red.ClusterOptions{
			Addrs:        strings.Split(r.Addr, ","),
			Password:     r.Pass,
			MaxRetries:   maxRetries,
			MinIdleConns: idleConns,
			TLSConfig:    tlsConfig,
			PoolSize:     r.PoolSize,
		})
		store.AddHook(SlowCallHook{})

		return store, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(*red.ClusterClient), nil
}
func delClusterClient(r *Redis) {
	clusterManager.Delete(r.Addr)
}

package lock

import (
	"fmt"
	"math/rand"
	"time"
	"unitechs.com/unios-dice/uni-base/lib/stores/redis"
)

var RedisClient *Lock

func AcquireLock() (bool, error) {
	if RedisClient != nil {
		return RedisClient.AcquireLock()
	}
	return false, fmt.Errorf("redisClient is nil")
}
func ReleaseLock(lockKey string) error {
	if RedisClient != nil {
		return RedisClient.ReleaseLockWithKey(lockKey)
	}
	return fmt.Errorf("redisClient is nil")
}

func InitRedisLock(rds *redis.Redis, lockKey string, lockTimeout int) (*Lock, error) {
	lock, err := NewRedisLock(rds, lockKey, lockTimeout)
	if err != nil {
		return nil, err
	}
	RedisClient = lock
	return lock, nil
}

func NewRedisLock(rds *redis.Redis, lockKey string, lockTimeout int) (*Lock, error) {
	if lockKey == "" {
		return nil, fmt.Errorf("lockKey is empty")
	}
	if rds == nil {
		return nil, fmt.Errorf("rds is nil")
	}
	if lockTimeout == 0 {
		lockTimeout = 2
	}

	return &Lock{
		rds:            rds,
		lockKey:        lockKey,
		lockTimeSecond: lockTimeout,
	}, nil
}

// Lock
// 这是一个全局锁，用于分布式锁的实现， 因为ctg-cache 不支持lua 脚本，这种实现方式在高并发的情况下，可能会存在问题
// 高并发的核心逻辑， 不能使用
type Lock struct {
	rds            *redis.Redis
	lockKey        string // redis 中存储锁的key值
	lockTimeSecond int    // 锁的过期时间
}

// AcquireLock 获取分布式锁，不使用Lua脚本
func (l *Lock) AcquireLock() (bool, error) {
	return l.AcquireLockWithKey(l.lockKey, l.lockTimeSecond)
}

// ReleaseLock 释放分布式锁，非原子操作（可能存在竞态条件）
func (l *Lock) ReleaseLock() error {
	return l.ReleaseLockWithKey(l.lockKey)
}

// AcquireLockWithKey 获取分布式锁，不使用Lua脚本
func (l *Lock) AcquireLockWithKey(lockKey string, lockTimeout int) (bool, error) {
	// 使用NX选项设置键值，只有键不存在时才设置成功
	// 不让同时获取锁
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)*2))
	success, err := l.rds.Setnx(lockKey, "1")
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if success {
		// 设置锁的过期时间以防止死锁
		expireErr := l.rds.Expire(lockKey, lockTimeout)
		if expireErr != nil {
			// 即使设置过期时间失败，锁已经被获取，需要手动处理这种情况
			_, _ = l.rds.Del(lockKey) // 可选：尝试删除已获取但未设置过期时间的锁
			return false, fmt.Errorf("acquired the lock but failed to set its expiration: %w", expireErr)
		}
		return true, nil
	}

	return false, nil
}

func (l *Lock) Flush() {
	_, _ = l.rds.Del(l.lockKey)
}

// ReleaseLockWithKey 释放分布式锁，非原子操作（可能存在竞态条件）
func (l *Lock) ReleaseLockWithKey(lockKey string) error {

	// 直接删除锁，但存在竞态条件，可能误删其他客户端的锁
	deleted, err := l.rds.Del(lockKey)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if deleted == 0 {
		// 锁可能已被其他客户端删除或者从未被当前客户端持有
		return fmt.Errorf("failed to release lock, key was not held by this client")
	}

	return nil
}

package conf

import (
	"context"
	"fmt"
	"time"

	redsyncredis "github.com/go-redsync/redsync/v4/redis"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

type RedisMutexLock string

const (
	ScanRMutex        = RedisMutexLock("scan")
	FetchRMutex       = RedisMutexLock("fetch")
	UserBalanceRMutex = RedisMutexLock("UB")
	UserApproveRMutex = RedisMutexLock("UA")
	LogProcessRMutex  = RedisMutexLock("LP")
	LogFlushRMutex    = RedisMutexLock("LF")
)

var RedisClient *RedisClientExtended // RedisUrl    string

type RedisClientExtended struct {
	redis.Client
	mutexes map[int64]map[RedisMutexLock]*redsync.Mutex
	pool    map[int64]redsyncredis.Pool
}

func LoadRedis() {
	time.Sleep(Config.RedisMongoSlowLoading)
	cl := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", Config.RedisUrl.Hostname(), Config.RedisUrl.Port()),
		DB:   Config.RedisDB,
	})
	RedisClient = &RedisClientExtended{
		*cl,
		make(map[int64]map[RedisMutexLock]*redsync.Mutex, 0),
		make(map[int64]redsyncredis.Pool, 0),
	}

	if _, err := RedisClient.GetOrSetTTL(context.Background(), "-cconn-", "-ok-", time.Second); err != nil {
		fmt.Println(err)
		Logger.Fatalf("RedisConnectionCheck: %+v", err)
	}
	if err := RedisClient.loadPools(); err != nil {
		fmt.Println(err)
		Logger.Fatalf("RedisConnectionCheck: %+v", err)
	}
	if err := RedisClient.loadMutexes(); err != nil {
		fmt.Println(err)
		Logger.Fatalf("RedisConnectionCheck: %+v", err)
	}
}

func (cl *RedisClientExtended) loadPools() error {
	for _, chain := range Config.SupportedChains {
		cl.pool[chain] = goredis.NewPool(&cl.Client)
	}
	return nil
}

func (cl *RedisClientExtended) loadMutexes() error {
	for chain, pool := range cl.pool {
		cl.mutexes[chain] = make(map[RedisMutexLock]*redsync.Mutex)
		rs := redsync.New(pool)
		cl.mutexes[chain][ScanRMutex] = rs.NewMutex(string(ScanRMutex))
	}
	return nil
}

func (r *RedisClientExtended) ChainMutex(chainId int64, key RedisMutexLock) *redsync.Mutex {
	return r.mutexes[chainId][key]
}

func (r *RedisClientExtended) GetOrSet(context context.Context, key string, value string) (string, error) {
	return r.GetOrSetTTL(context, key, value, redis.KeepTTL)
}

func (r *RedisClientExtended) GetOrSetTTL(
	context context.Context, key string, value string, ttl time.Duration,
) (string, error) {
	if res := r.Get(context, key); res.Err() != nil {
		if res.Err() == redis.Nil {
			if res := r.Set(context, key, value, ttl); res.Err() != nil {
				return "", res.Err()
			}
		} else {
			return "", res.Err()
		}
	} else {
		value = res.Val()
	}
	return value, nil
}

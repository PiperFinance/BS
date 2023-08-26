package conf

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RedisMutexLock string

const (
	ScanRMutex        = RedisMutexLock("scan")
	FetchRMutex       = RedisMutexLock("fetch")
	UserBalanceRMutex = RedisMutexLock("UB")
	UserApproveRMutex = RedisMutexLock("UA")
	LogProcessRMutex  = RedisMutexLock("LP")
	LogFlushRMutex    = RedisMutexLock("LF")
	VaccumeBlockRKey  = "BS:VB:%d"
	VaccumeObjIDRKey  = "BS:VID:%d"
)

var (
	RedisClient       *RedisClientExtended // RedisUrl    string
	vaccumeObjIdMutex = sync.Mutex{}
)

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
		Logger.Panicf("RedisConnectionCheck: %+v", err)
	}
	if err := RedisClient.loadPools(); err != nil {
		fmt.Println(err)
		Logger.Panicf("RedisConnectionCheck: %+v", err)
	}
	if err := RedisClient.loadMutexes(); err != nil {
		fmt.Println(err)
		Logger.Panicf("RedisConnectionCheck: %+v", err)
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

func (r *RedisClientExtended) IncrHSet(context context.Context, key string, field string) error {
	var val int64
	if cmd := r.HGet(context, key, field); cmd.Err() != nil && cmd.Err() != redis.Nil {
		return cmd.Err()
	} else {
		val, _ = strconv.ParseInt(cmd.Val(), 10, 64)
	}

	cmd := r.HSet(context, key, field, val+1)
	return cmd.Err()
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

type logIdVaccum struct {
	Ids []primitive.ObjectID `json:"ids"`
}

func (r *RedisClientExtended) SetParsedLogsIDsToVaccum(ctx context.Context, chain int64, ObjIds []primitive.ObjectID) error {
	k := fmt.Sprintf(VaccumeObjIDRKey, chain)
	ids := logIdVaccum{Ids: ObjIds}
	if cmd := r.Get(ctx, k); cmd.Err() == nil {
		prevIds := logIdVaccum{}
		if err := json.Unmarshal([]byte(cmd.Val()), &prevIds); err != nil {
			return err
		}
		ids.Ids = append(ids.Ids, prevIds.Ids...)
	}
	val, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	res := r.Set(ctx, k, val, -1)
	return res.Err()
}

func (r *RedisClientExtended) GetParsedLogsIDsToVaccum(ctx context.Context, chain int64) ([]primitive.ObjectID, error) {
	k := fmt.Sprintf(VaccumeObjIDRKey, chain)
	if cmd := r.Get(ctx, k); cmd.Err() == nil {
		prevIds := logIdVaccum{}
		if err := json.Unmarshal([]byte(cmd.Val()), &prevIds); err != nil {
			return nil, err
		}
		return prevIds.Ids, nil
	} else if cmd.Err() == redis.Nil {
		return nil, nil
	} else {
		return nil, cmd.Err()
	}
}

type VaccumBlockRange struct {
	FromBlock uint64 `json:"fb"`
	ToBlock   uint64 `json:"tb"`
}

// SetRawLogsToVaccum Adds Block Range for later vaccum, corresponds to PasredLogs collection
func (r *RedisClientExtended) SetRawLogsToVaccum(ctx context.Context, chain int64, fromBlock uint64, toBlock uint64) error {
	k := fmt.Sprintf(VaccumeBlockRKey, chain)
	vacRng := VaccumBlockRange{FromBlock: fromBlock, ToBlock: toBlock}
	val, err := json.Marshal(vacRng)
	if err != nil {
		return err
	}
	res := r.RPush(ctx, k, val)
	return res.Err()
}

// GetRawLogsToVaccum Block Range vaccum, corresponds to PasredLogs collection
// stop at (nil, nil) response
func (r *RedisClientExtended) GetRawLogsToVaccum(ctx context.Context, chain int64) (*VaccumBlockRange, error) {
	k := fmt.Sprintf(VaccumeBlockRKey, chain)
	vacRng := VaccumBlockRange{}
	if cmd := r.LPop(ctx, k); cmd.Err() == nil {
		if err := json.Unmarshal([]byte(cmd.Val()), &vacRng); err != nil {
			return nil, err
		}
	} else if cmd.Err() == redis.Nil {
		return nil, nil
	}
	return &vacRng, nil
}

// ReentrancyCheck Returns ok if no previous record is found / err if redis say so =)
func (r *RedisClientExtended) ReentrancyCheck(ctx context.Context, chainId int64, field string) error {
	k := fmt.Sprintf("BS:RC:%d", chainId)
	if cmd := r.HExists(ctx, k, field); cmd.Err() == redis.Nil || (cmd.Err() == nil && !cmd.Val()) {
		// TODO: Check err
		r.HSet(ctx, k, field, true)
		return nil
	} else if cmd.Err() == nil {
		if !cmd.Val() {
			return nil
		} else {
			return fmt.Errorf("Tried to re enter using key %s ", field)
		}
	} else {
		return cmd.Err()
	}
}

// UserTokenHSKey Hash Set containing user's token + balance in each chain
func UserTokenHSKey(chain int64, user common.Address, token common.Address) (string, string) {
	return fmt.Sprintf("BS:UTHS:%d", chain), fmt.Sprintf("%s-%s", user.String(), token.String())
}

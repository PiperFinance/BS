package utils

import (
	"bytes"
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/PiperFinance/BS/src/conf"
)

func IsRegistered(add common.Address) bool {
	found, ok := conf.OnlineUsers.AllAdd[add]
	return ok && found
}

func IsAddressNull(ctx context.Context, chainId int64, user common.Address) bool {
	return user.Big().Cmp(big.NewInt(0)) < 1
}

func IsAddressToken(ctx context.Context, chainId int64, user common.Address) bool {
	if res := conf.GetMongoCol(chainId, conf.TokenColName).FindOne(ctx, bson.D{{Key: "_id", Value: user}}); res.Err() == mongo.ErrNoDocuments {
		return false
	} else if res.Err() == nil {
		return true
	} else {
		conf.Logger.Errorw("IsAddressAToken", "err", res.Err(), "user", user, "chain", chainId)
		return false
	}
}

func IsUserBanned(ctx context.Context, chainId int64, user common.Address) bool {
	if res := conf.GetMongoCol(chainId, conf.BannedUsersColName).FindOne(ctx, bson.D{{Key: "_id", Value: user.String()}}); res.Err() == mongo.ErrNoDocuments {
		return false
	} else if res.Err() == nil {
		return true
	} else {
		conf.Logger.Errorw("BannedUsers", "err", res.Err(), "user", user, "chain", chainId)
		return false
	}
}

func IsDuplicated(user common.Address, token common.Address) bool {
	return bytes.Equal(user.Bytes(), token.Bytes())
}

func IsLimited(ctx context.Context, chainId int64, user common.Address, token common.Address) bool {
	// FIXME - change this back if you see huge number of requests happening !
	return IsDuplicated(user, token) || IsAddressNull(ctx, chainId, user) || IsAddressToken(ctx, chainId, user) || IsUserBanned(ctx, chainId, user)
}

func AddNew(ctx context.Context, chainId int64, user common.Address, token common.Address) error {
	k, f := conf.UserTokenHSKey(chainId, user, token)
	return conf.RedisClient.HSet(ctx, k, f, true).Err()
}

func IsNew(ctx context.Context, chainId int64, user common.Address, token common.Address) (error, bool) {
	// NOTE - For Request CountReduction contracts contract and zero address is not included
	if (conf.Config.LimitUsers && !conf.OnlineUsers.IsAddressOnline(user)) || IsLimited(ctx, chainId, user, token) {
		return nil, false
	}

	k, f := conf.UserTokenHSKey(chainId, user, token)
	if cmd := conf.RedisClient.HExists(ctx, k, f); cmd.Err() != nil {
		if cmd.Err() == redis.Nil {
			return nil, false
		} else {
			return cmd.Err(), false
		}
	} else {
		return nil, !cmd.Val()
	}
}

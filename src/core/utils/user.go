package utils

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/PiperFinance/BS/src/conf"
)

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
	if res := conf.GetMongoCol(chainId, conf.UserBalColName).FindOne(ctx, bson.D{{Key: "_id", Value: user.String()}}); res.Err() == mongo.ErrNoDocuments {
		return false
	} else if res.Err() == nil {
		return true
	} else {
		conf.Logger.Errorw("BannedUsers", "err", res.Err(), "user", user, "chain", chainId)
		return false
	}
}

func IsLimited(ctx context.Context, chainId int64, user common.Address) bool {
	return IsAddressNull(ctx, chainId, user) || IsAddressToken(ctx, chainId, user) || IsUserBanned(ctx, chainId, user)
}

func IsNew(ctx context.Context, chainId int64, user common.Address, token common.Address) (error, bool) {
	// TODO - Add user Limit here
	// FIXME - For Request CountReduction contracts contract and zero address is not included
	if conf.Config.LimitUsers && !conf.OnlineUsers.IsAddressOnline(user) {
		return nil, false
	}
	if IsLimited(ctx, chainId, user) {
		return nil, false
	}

	filter := bson.D{{Key: "user", Value: user}, {Key: "token", Value: token}}
	if count, err := conf.GetMongoCol(chainId, conf.UserBalColName).CountDocuments(ctx, filter); count == 0 || err == mongo.ErrNoDocuments {
		return nil, true
	} else {
		if err == nil {
			conf.Logger.Infow("NewUserFinder", "user", user, "token", token, "err", err)
		} else {
			conf.Logger.Errorw("NewUserFinder", "user", user, "token", token, "err", err)
		}
		return err, false
	}
}

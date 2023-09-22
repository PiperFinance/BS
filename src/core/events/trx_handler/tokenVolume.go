package trx_handler

import (
	"context"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// updateTokens goes over tokens and makes all EmitterAddress(tokens) are stored in db
func (h *UserTrxHandler) updateTokens(ctx context.Context, chainId int64, transfers []*schema.LogTransfer) error {
	// TODO: calculate token's transfer volume here as well
	col := conf.GetMongoCol(chainId, conf.TokenColName)
	uniqueTokens := make([]common.Address, 0)
	var tokenExists bool
	for i, transfer := range transfers {
		if transfer == nil {
			conf.Logger.Error(i)
			continue
		}
		_token := transfer.EmitterAddress
		tokenExists = true
		for _, token := range uniqueTokens {
			if token == _token {
				tokenExists = false
				break
			}
		}
		if tokenExists {
			uniqueTokens = append(uniqueTokens, _token)
		}
	}
	for _, token := range uniqueTokens {
		if count, err := col.CountDocuments(ctx, bson.D{{Key: "_id", Value: token}}); count == 0 || err == mongo.ErrNoDocuments {
			// tokens = append(tokens, )
			// TODO: - check err later
			col.InsertOne(ctx, bson.D{{Key: "_id", Value: token}})
		} else if err != nil {
			return err
		}
	}
	return nil
}

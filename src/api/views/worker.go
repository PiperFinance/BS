package views

import (
	"strconv"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func LastScannedBlock(c *fiber.Ctx) error {
	// TODO - FIX this to be multi chain ...
	chainQ := c.Query("chain", "")
	chain, err := strconv.ParseInt(chainQ, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"err": "chain field should be an int!"})
	}
	lastBlock, err := utils.GetLastBlock(chain)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err.Error(),
		})
	} else {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"res": lastBlock,
		})
	}
}

func LastScannedBlocks(c *fiber.Ctx) error {
	chainQ := c.Query("chain", "")
	chain, err := strconv.ParseInt(chainQ, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"err": "chain field should be an int!"})
	}

	cursor, err := conf.GetMongoCol(chain, conf.BlockColName).Find(
		c.Context(),
		bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err.Error(),
		})
	} else {
		r := make([]schema.BlockM, 100)
		i := 0
		for i < 100 && cursor.Next(c.Context()) {
			err = cursor.Decode(&r[i])
			if err != nil {
				conf.Logger.Errorf("Decoding Block Failed @ %s", err)
				continue
			}
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"res": r,
		})
	}
}

func CallStatus(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg": conf.CallCount,
	})
}

// func MissedBlocks(c *fiber.Ctx) error {
// 	var lastBlock uint64
// 	if res := conf.RedisClient.Get(ctx, tasks.LastScannedBlockKey); res.Err() == nil{
// 		lb , err := res.Uint64()
// 		if err != nil {
// 			return  err
// 		}else{
// 			lastBlock= lb
// 		}
// 	}

// 	// for conf.Config.StartingBlockNumber
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"res": lastBlock,
// 	})
// }

package views

import (
	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/utils"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func LastScannedBlock(c *fiber.Ctx) error {
	lastBlock, err := utils.GetLastBlock()
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
	cursor, err := conf.MongoDB.Collection(conf.BlockColName).Find(
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
				log.Errorf("Decoding Block Failed @ %s", err)
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

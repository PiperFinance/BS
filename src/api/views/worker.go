package views

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/utils"
)

func LastScannedBlock(c *fiber.Ctx) error {
	chainQ := c.Query("chain", "")
	if len(chainQ) > 0 {
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
	} else {
		lastBlocks := make(map[int64]uint64, len(conf.SupportedNetworks))
		for _, chain := range conf.Config.SupportedChains {
			if lastBlock, err := utils.GetLastBlock(chain); err != nil {
				lastBlocks[chain] = 0
			} else {
				lastBlocks[chain] = lastBlock
			}
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"res": lastBlocks,
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

func Status(c *fiber.Ctx) error {
	r := make(map[int64]map[string]string)
	for _, chain := range conf.Config.SupportedChains {
		r[chain] = make(map[string]string)
		r[chain]["Call"] = conf.CallCount.StatusChain(chain)
		r[chain]["FailedCalls"] = conf.FailedCallCount.StatusChain(chain)
		r[chain]["MultiCall"] = conf.MultiCallCount.StatusChain(chain)
		r[chain]["NewUsers"] = conf.NewUsersCount.StatusChain(chain)
		r[chain]["NewBlock"] = conf.NewBlockCount.StatusChain(chain)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg": r,
	})
}

func CallStatus(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg": conf.CallCount,
	})
}

func NewBlockStatusSimple(c *fiber.Ctx) error {
	r := make(map[int64]string)
	for _, chain := range conf.Config.SupportedChains {
		r[chain] = conf.NewBlockCount.StatusChain(chain)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg": r,
	})
}

func NewBlockStatus(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg": conf.NewBlockCount,
	})
}

type Stats struct {
	Scanned int64 `json:"scanned"`
	Fetched int64 `json:"fetched"`
	Added   int64 `json:"added"`
	Parsed  int64 `json:"parsed"`
}

func BlockStats(c *fiber.Ctx) error {
	r := make(map[int64]Stats)
	for _, chain := range conf.Config.SupportedChains {

		scanned, err := conf.GetMongoCol(chain, conf.BlockColName).CountDocuments(c.Context(), bson.M{"status": schema.Scanned})
		if err != nil {
			conf.Logger.Errorf("Scanned Blocks on chain %d : %+v", chain, err)
		}
		fetched, _ := conf.GetMongoCol(chain, conf.BlockColName).CountDocuments(c.Context(), bson.M{"status": schema.Fetched})
		parsed, _ := conf.GetMongoCol(chain, conf.BlockColName).CountDocuments(c.Context(), bson.M{"status": schema.Parsed})
		added, _ := conf.GetMongoCol(chain, conf.BlockColName).CountDocuments(c.Context(), bson.M{"status": schema.Added})
		r[chain] = Stats{
			Scanned: scanned,
			Fetched: fetched,
			Added:   added,
			Parsed:  parsed,
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg": r,
	})
}

func MissedBlocks(c *fiber.Ctx) error {
	r := make(map[int64][]schema.BlockM)
	for _, chain := range conf.Config.SupportedChains {
		block, err := conf.LatestBlock(c.Context(), chain)
		if err != nil {
			return err
		}
		r[chain] = make([]schema.BlockM, 0)
		col := conf.GetMongoCol(chain, conf.BlockColName)
		filter := bson.M{"status": bson.D{{Key: "$ne", Value: schema.Added}}, "no": bson.D{{Key: "$lt", Value: block - conf.Config.BlockHeadDelay}}}
		if curs, err := col.Find(c.Context(), filter); err != nil {
			return err
		} else {
			for curs.Next(c.Context()) {
				ub := schema.BlockM{}
				err := curs.Decode(&ub)
				if err != nil {
					conf.Logger.Errorf("GetBal: %s", err.Error())
					continue
				}
				r[chain] = append(r[chain], ub)
			}
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg": r,
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

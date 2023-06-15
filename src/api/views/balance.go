package views

import (
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	coreUtils "github.com/PiperFinance/BS/src/core/utils"
)

func GetBal(c *fiber.Ctx) error {
	user := c.Query("user", "")
	if !common.IsHexAddress(user) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"err": "user field in required and should be in formate of 0x...!"})
	}
	chainQ := c.Query("chain", "")
	chain, err := strconv.ParseInt(chainQ, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"err": "chain field should be an int!"})
	}
	add := common.HexToAddress(user)

	curs, err := conf.GetMongoCol(chain, conf.UserBalColName).Find(c.Context(), bson.M{
		"user": add,
	})
	if err != nil {
		return err
	}
	r := make([]schema.UserBalance, 0)
	for curs.Next(c.Context()) {
		ub := schema.UserBalance{}
		err := curs.Decode(&ub)
		if err != nil {
			conf.Logger.Errorf("GetBal: %s", err.Error())
		}
		r = append(r, ub)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"res": r,
	})
}

func SetBal(c *fiber.Ctx) error {
	chain, err := strconv.ParseInt(c.Params("chain"), 10, 64)
	if err != nil {
		return c.SendStatus(422)
	}
	var payload []schema.UserBalance
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	col := conf.GetMongoCol(chain, conf.UserBalColName)
	userTokens := make([]interface{}, 0)
	currentBlock, err := conf.EthClient(chain).BlockNumber(c.Context())
	// conf.CallCount.Add(chain)
	if err != nil {
		// TODO - change err type
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}
	for _, userBal := range payload {
		if err, ok := coreUtils.IsNew(c.Context(), chain, userBal.User, userBal.Token); err == nil && ok {
			userBal.ChangedAt = currentBlock
			userBal.StartedAt = currentBlock
			userTokens = append(userTokens, userBal)
		} else if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "message": err.Error()})
		}
	}
	if _, err := col.InsertMany(c.Context(), userTokens); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}
	return nil
}

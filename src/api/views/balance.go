package views

import (
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	coreUtils "github.com/PiperFinance/BS/src/core/utils"
)

func GetBal(c *fiber.Ctx) error {
	token := c.Query("token", "")
	if len(token) > 0 && !common.IsHexAddress(token) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"err": "token field in required and should be in formate of 0x...!"})
	}
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
	var filter bson.M
	if len(token) > 0 {
		filter = bson.M{
			"user":  add,
			"token": common.HexToAddress(token),
		}
	} else {
		filter = bson.M{
			"user": add,
		}
	}
	curs, err := conf.GetMongoCol(chain, conf.UserBalColName).Find(c.Context(), filter)
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
	// TODO: check token , user common add binaries
	var payload []schema.UserBalance
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	col := conf.GetMongoCol(chain, conf.UserBalColName)
	userTokens := make([]interface{}, 0)
	currentBlock, err := conf.EthClient(chain).BlockNumber(c.Context())
	conf.CallCount.Add(chain)
	if err != nil {
		conf.FailedCallCount.Add(chain)
		// TODO: - change err type
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}
	for _, userBal := range payload {
		if userBal.Token == conf.NetworkValueAddress(chain) {
			continue
		}
		if err, ok := coreUtils.IsNew(c.Context(), chain, userBal.User, userBal.Token); err == nil && ok {
			userBal.ChangedAt = currentBlock
			userBal.StartedAt = currentBlock
			userTokens = append(userTokens, userBal)
			coreUtils.AddNew(c.Context(), chain, userBal.User, userBal.Token)
		} else if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "message": err.Error()})
		}
	}
	if len(userTokens) > 0 {
		if _, err := col.InsertMany(c.Context(), userTokens); err != nil && !strings.Contains(err.Error(), "duplicate") {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "message": err.Error()})
		}
	}
	return c.SendStatus(fiber.StatusAccepted)
}

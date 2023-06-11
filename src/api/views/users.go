package views

import (
	"strconv"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func GetUsers(c *fiber.Ctx) error {
	token := c.Query("token", "")
	if !common.IsHexAddress(token) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"err": "user field in required and should be in formate of 0x...!"})
	}
	chainQ := c.Query("chain", "")
	chain, err := strconv.ParseInt(chainQ, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"err": "chain field should be an int!"})
	}

	add := common.HexToAddress(token)
	curs, err := conf.GetMongoCol(chain, conf.UserBalColName).Find(c.Context(), bson.M{
		"token": add,
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

type UserRequest struct {
	Users  []common.Address `json:"users"`
	Tokens []common.Address `json:"tokens"`
	Chains []int64          `json:"chains"`
}

func GetUser(c *fiber.Ctx) error {
	r := UserRequest{}
	if err := c.QueryParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"err": err.Error()})
	}
	res := make(map[int64][]interface{}, len(r.Chains))
	filter := bson.M{"user": r.Users[0]}
	for _, chain := range r.Chains {
		if count, err := conf.GetMongoCol(chain, conf.UserBalColName).CountDocuments(c.Context(), filter); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"err": err.Error()})
		} else {
			if count == 0 {
				continue
			}
			res[chain] = make([]interface{}, count)
		}
		if cursor, err := conf.GetMongoCol(chain, conf.UserBalColName).Find(
			c.Context(), filter); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"err": err.Error()})
		} else {
			cursor.All(c.Context(), res[chain])
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"res": res})
}

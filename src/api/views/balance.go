package views

import (
	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/charmbracelet/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func GetBal(c *fiber.Ctx) error {
	user := c.Query("user", "")
	if !common.IsHexAddress(user) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"err": "user field in required and should be in formate of 0x...!"})
	}
	add := common.HexToAddress(user)
	curs, err := conf.MongoDB.Collection(conf.UserBalColName).Find(c.Context(), bson.M{
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
			log.Errorf("GetBal: %s", err.Error())
		}
		r = append(r, ub)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"res": r,
	})
}

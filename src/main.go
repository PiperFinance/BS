package main

import (
	"context"

	"github.com/PiperFinance/BS/src/conf"
	contract_helpers "github.com/PiperFinance/BS/src/contracts/helpers"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/joho/godotenv/autoload"
)

func init() {
	conf.LoadConfig()
	conf.LoadLogger()
	conf.LoadMongo()
	conf.LoadRedis()
	conf.LoadMainNets()
	conf.LoadNetwork()
	conf.LoadQueue()
	conf.LoadDebugItems()
}

// ONLY FOR TESTING PURPOSES ...

func main() {
	usersTokens := []contract_helpers.UserToken{
		{User: common.HexToAddress("0x02631bb2d276af822aee9d02ff5bd5b5edfa4994"), Token: common.HexToAddress("0x049d68029688eabf473097a2fc38ef61633a3c7a")},
	}
	bal := contract_helpers.EasyBalanceOf{UserTokens: usersTokens, ChainId: 250, BlockNumber: 58545713}
	if err := bal.Execute(context.Background()); err != nil {
		conf.Logger.Fatal(err)
	} else {
		conf.Logger.Info(bal.UserTokens[0].Balance.String())
		conf.Logger.Info(bal.UserTokens[0].User.String())
		conf.Logger.Info(bal.UserTokens[0].Token.String())
	}
	// (&StartConf{}).StartAll()
	// select {}
}

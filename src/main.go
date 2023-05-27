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
		{User: common.HexToAddress("0xabca9665c76722b0b40643ca38672915bd259476"), Token: common.HexToAddress("0x5f7f94a1dd7b15594d17543beb8b30b111dd464c")},
	}
	bal := contract_helpers.EasyBalanceOf{UserTokens: usersTokens, ChainId: 250, BlockNumber: 63200835}
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

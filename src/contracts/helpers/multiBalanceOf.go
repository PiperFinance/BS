package contract_helpers

import (
	"context"
	"fmt"
	"math/big"

	"github.com/PiperFinance/BS/src/conf"
	Multicall "github.com/PiperFinance/BS/src/contracts/MulticallContract"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

var (
	ERC20_BALANCE_OF_FUNC = "70a08231" //	balanceOf(address)
	MULTICALL_V3_ADDRESS  = common.HexToAddress("0xca11bde05977b3631167028862be2a173976ca11")
)

type UserToken struct {
	User    common.Address
	Token   common.Address
	Balance *big.Int
	call    Multicall.Multicall3Call3
}

type EasyBalanceOf struct {
	// Tokens      []common.Address
	// Users       []common.Address
	UserTokens  []UserToken
	BlockNumber int64
	ChainId     int64
}

// ParseBigIntResult If there is only one bigint in call's response
func ParseBigIntResult(result []byte) *big.Int {
	z := big.NewInt(0)
	if len(result) > 32 {
		z.SetBytes(result[:32])
	} else {
		z.SetBytes(result)
	}
	return z

	// if z.Cmp(big.NewInt(100)) <= 0 {
	// 	return z
	// } else {
	// 	return z
	// }
}

func BalanceOf(call UserToken) Multicall.Multicall3Call3 {
	return Multicall.Multicall3Call3{
		Target:       call.Token,
		AllowFailure: true,
		CallData:     common.Hex2Bytes(fmt.Sprintf("%s%s", ERC20_BALANCE_OF_FUNC, call.User.Hash().String()[2:])),
	}
}

func (bal *EasyBalanceOf) populateTokenBalanceCalls() {
	for i, userToken := range bal.UserTokens {
		bal.UserTokens[i].call = BalanceOf(userToken)
	}
}

func (self *EasyBalanceOf) multiCaller() *Multicall.MulticallCaller {
	contractInstance, err := Multicall.NewMulticallCaller(MULTICALL_V3_ADDRESS, conf.EthClient(self.ChainId))
	if err != nil {
		conf.Logger.Fatalf("Multicall Contract Gen : %+v", err)
	}
	return contractInstance
}

func (self *EasyBalanceOf) Execute(ctx context.Context) error {
	self.populateTokenBalanceCalls()
	calls := make([]Multicall.Multicall3Call3, len(self.UserTokens))
	for i, userTokens := range self.UserTokens {
		calls[i] = userTokens.call
	}

	ctxWTimeout, _ := context.WithTimeout(ctx, conf.Config.MultiCallTimeout)
	var cOpts bind.CallOpts
	if self.BlockNumber > 1 {
		cOpts = bind.CallOpts{Context: ctxWTimeout, BlockNumber: big.NewInt(self.BlockNumber)}
	} else {
		cOpts = bind.CallOpts{Context: ctxWTimeout}
	}

	for i, _call := range calls {
		conf.Logger.Infof("[%d][%s][%s]", i, _call.Target, common.Bytes2Hex(_call.CallData))
	}

	res, err := self.multiCaller().Aggregate3(&cOpts, calls)

	if err != nil {
		return err
	} else {
		for i, _res := range res {
			if _res.Success {
				self.UserTokens[i].Balance = ParseBigIntResult(_res.ReturnData)
			}
		}
	}
	return nil
}
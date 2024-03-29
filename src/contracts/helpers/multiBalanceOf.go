package contract_helpers

import (
	"context"
	"fmt"
	"math/big"

	"github.com/PiperFinance/BS/src/conf"
	Multicall "github.com/PiperFinance/BS/src/contracts/MulticallContract"
	"github.com/PiperFinance/BS/src/utils"
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

func (easBal *EasyBalanceOf) multiCaller() *Multicall.MulticallCaller {
	contractInstance, err := Multicall.NewMulticallCaller(MULTICALL_V3_ADDRESS, conf.EthClient(easBal.ChainId))
	if err != nil {
		conf.Logger.Panicf("Multicall Contract Gen : %+v", err)
	}
	return contractInstance
}

func (easBal *EasyBalanceOf) Execute(ctx context.Context) error {
	easBal.populateTokenBalanceCalls()
	calls := make([]Multicall.Multicall3Call3, len(easBal.UserTokens))
	for i, userTokens := range easBal.UserTokens {
		calls[i] = userTokens.call
	}

	ctxWTimeout, cancel := context.WithTimeout(ctx, conf.Config.MultiCallTimeout)
	defer cancel()
	var cOpts bind.CallOpts
	if easBal.BlockNumber > 1 {
		cOpts = bind.CallOpts{Context: ctxWTimeout, BlockNumber: big.NewInt(easBal.BlockNumber)}
	} else {
		cOpts = bind.CallOpts{Context: ctxWTimeout}
	}

	res, err := easBal.multiCaller().Aggregate3(&cOpts, calls)
	conf.CallCount.Add(easBal.ChainId)
	if err != nil {
		return &utils.RpcError{Err: err, ChainId: easBal.ChainId, BlockNumber: uint64(easBal.BlockNumber), Name: "MultiCall"}
	} else {
		for i, _res := range res {
			// NOTE: weird panic case :|
			if i >= len(easBal.UserTokens) {
				continue
			}
			if _res.Success {
				easBal.UserTokens[i].Balance = ParseBigIntResult(_res.ReturnData)
			} else {
				if !conf.Config.SilenceMulticallErrs {
					conf.Logger.Errorw("Multicall", "res", _res.ReturnData, "chain", easBal.ChainId, "block", easBal.BlockNumber, "user", easBal.UserTokens[i].User.String(), "token", easBal.UserTokens[i].Token.String())
				}
				easBal.UserTokens[i].Balance = big.NewInt(0)
			}
		}
	}
	return nil
}

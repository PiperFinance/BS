package schema

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type User struct {
	common.Address
}

type UserBalance struct {
	TokenStr  string         `bson:"tokenStr" json:"tokenStr"`
	UserStr   string         `bson:"userStr" json:"userStr"`
	User      common.Address `bson:"user" json:"user"`
	Token     common.Address `bson:"token" json:"token"`
	TrxCount  uint64         `bson:"count" json:"count"`
	ChangedAt uint64         `bson:"c_t" json:"c_t"`
	StartedAt uint64         `bson:"s_t" json:"s_t"`
	Balance   string         `bson:"bal" json:"bal"`
}

func (ub *UserBalance) SetBalanceInt(newBal int64) {
	ub.SetBalance(big.NewInt(newBal))
}

func (ub *UserBalance) SetBalance(newBal *big.Int) {
	ub.Balance = newBal.String()
}

func (ub *UserBalance) GetBalance() (*big.Int, bool) {
	v := big.Int{}
	return v.SetString(ub.Balance, 10)
}

func (ub *UserBalance) GetBalanceStr() string {
	return ub.Balance
}

func (ub *UserBalance) AddBal(b *big.Int) error {
	if b == nil {
		return fmt.Errorf("b is nil ")
	}
	a, ok := ub.GetBalance()
	if !ok {
		return fmt.Errorf("failed to cast %s to big.Int", ub.Balance)
	}
	ub.SetBalance(a.Add(a, b))
	return nil
}

func (ub *UserBalance) SubBal(b *big.Int) error {
	if b == nil {
		return fmt.Errorf("b is nil ")
	}
	a, ok := ub.GetBalance()
	if !ok {
		return fmt.Errorf("failed to cast %s to big.Int", ub.Balance)
	}
	ub.SetBalance(a.Sub(a, b))
	return nil
}

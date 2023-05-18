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
	User      common.Address `bson:"user" json:"user"`
	Token     common.Address `bson:"token" json:"token"`
	ChangedAt uint64         `bson:"c_t" json:"c_t"`
	StartedAt uint64         `bson:"s_t" json:"s_t"`
	balance   string         `bson:"bal" json:"bal"`
}

func (ub *UserBalance) SetBalanceInt(newBal int64) {
	ub.SetBalance(big.NewInt(newBal))
}

func (ub *UserBalance) SetBalance(newBal *big.Int) {
	ub.balance = newBal.String()
}

func (ub *UserBalance) GetBalance() (*big.Int, bool) {
	v := big.Int{}
	return v.SetString(ub.balance, 10)
}

func (ub *UserBalance) GetBalanceStr() string {
	return ub.balance
}

func (ub *UserBalance) SubBal(b *big.Int) error {
	if b == nil {
		return fmt.Errorf("b is nil ")
	}
	a, ok := ub.GetBalance()
	if !ok {
		return fmt.Errorf("failed to cast %s to big.Int", ub.balance)
	}
	ub.SetBalance(a.Sub(a, b))
	return nil
}

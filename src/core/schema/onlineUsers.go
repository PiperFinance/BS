package schema

import (
	"github.com/ethereum/go-ethereum/common"
)

type OnlineUsers struct {
	AllAdd map[common.Address]bool
	// OnlineAdd  map[common.Address]bool
	// OnlineUser map[uuid.UUID]bool
}

func (ou *OnlineUsers) NewAdd(OnlineAddress common.Address) {
	ou.AllAdd[OnlineAddress] = true
}

func (ou *OnlineUsers) Refresh(OnlineAddress []common.Address) {
	for _, add := range OnlineAddress {
		ou.AllAdd[add] = true
		// ou.OnlineAdd[add] = true
	}
}

func (ou *OnlineUsers) IsAddressOnline(add common.Address) bool {
	_, ok := ou.AllAdd[add]
	return ok
}

package schema

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type OnlineUsers struct {
	OnlineUserMutex sync.Mutex
	AllAdd          map[common.Address]bool
	// OnlineAdd  map[common.Address]bool
	// OnlineUser map[uuid.UUID]bool
}

func (ou *OnlineUsers) NewAdd(OnlineAddress common.Address) {
	ou.OnlineUserMutex.Lock()
	ou.AllAdd[OnlineAddress] = true
	ou.OnlineUserMutex.Unlock()
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

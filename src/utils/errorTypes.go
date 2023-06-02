package utils

import "fmt"

type RpcError struct {
	Name        string
	Err         error
	RPC         string
	BlockNumber uint64 `bson:"no" json:"no"`
	ChainId     int64  `bson:"chain" json:"chain"`
}

func (e *RpcError) Error() string {
	return fmt.Sprintf("C_ID=%d , RPC=%s block_no=%d @%s err=%v", e.ChainId, e.RPC, e.BlockNumber, e.Name, e.Err)
}

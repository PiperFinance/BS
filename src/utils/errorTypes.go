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

type ErrEventParserNotFound struct {
	Event       string
	BlockNumber uint64 `bson:"no" json:"no"`
	TrxIndex    uint   `bson:"index" json:"index"`
	ChainId     int64  `bson:"chain" json:"chain"`
}

func (e *ErrEventParserNotFound) Error() string {
	return fmt.Sprintf("EventParserNotFound: event-hash=%s block=%d TrxIndex=%d ChainId=%d", e.Event, e.BlockNumber, e.TrxIndex, e.ChainId)
}

type ErrEventNoInput struct {
	Event       string
	BlockNumber uint64 `bson:"no" json:"no"`
	TrxIndex    uint   `bson:"index" json:"index"`
	ChainId     int64  `bson:"chain" json:"chain"`
}

func (e *ErrEventNoInput) Error() string {
	return fmt.Sprintf("EventParserNotFound: event-hash=%s block=%d TrxIndex=%d ChainId=%d", e.Event, e.BlockNumber, e.TrxIndex, e.ChainId)
}

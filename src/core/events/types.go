package events

import "context"

type (
	EventGroup     string
	EventSignature string
	EventName      string
)

const (
	TransferGrp = EventGroup("Transfer")
	ApprovalGrp = EventGroup("Approval")
)

// Handler Acts on what should be done to given event (i.e. saving in DB)
type Handler interface {
	Handle(interface{}) error
	Flush(ctx context.Context, chainId int64, blockNumber uint64) error
}

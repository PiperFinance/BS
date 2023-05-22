package schema

import "time"

type BlockTask struct {
	BlockNumber uint64
}

// MBlock Block Stored At Mongo
type BlockM struct {
	ScannerStatus `bson:"status"`
	BlockNumber   uint64    `bson:"no"`
	UpdatedAt     time.Time `bson:"c_at"`
	StartedAt     time.Time `bson:"s_at"`
}

func (bm *BlockM) Scan(t BlockTask) BlockM {
	return BlockM{
		ScannerStatus: Scanned,
		BlockNumber:   t.BlockNumber,
		UpdatedAt:     time.Now(),
		StartedAt:     time.Now(),
	}
}

package schema

import (
	"time"
)

type BatchBlockTask struct {
	FromBlockNumber uint64 `bson:"from_no"`
	ToBlockNumber   uint64 `bson:"to_no"`
	BlockNumber     uint64 `bson:"no"`
	ChainId         int64  `bson:"chain"`
}

type BlockTask struct {
	BlockNumber uint64 `bson:"no"`
	ChainId     int64  `bson:"chain"`
}

// MBlock Block Stored At Mongo
type BlockM struct {
	ScannerStatus string    `bson:"status"`
	BlockNumber   uint64    `bson:"no"`
	UpdatedAt     time.Time `bson:"c_at"`
	StartedAt     time.Time `bson:"s_at"`
	ChainId       int64     `bson:"chain"`
}

func (bm *BlockM) SetScanned() *BlockM {
	bm.ScannerStatus = Scanned
	bm.UpdatedAt = time.Now()
	bm.StartedAt = time.Now()
	return bm
}

func (bm *BlockM) SetFetched() *BlockM {
	bm.ScannerStatus = Fetched
	bm.UpdatedAt = time.Now()
	return bm
}

func (bm *BlockM) SetParsed() *BlockM {
	bm.ScannerStatus = Parsed
	bm.UpdatedAt = time.Now()
	return bm
}

func (bm *BlockM) SetAdded() *BlockM {
	bm.ScannerStatus = Added
	bm.UpdatedAt = time.Now()
	return bm
}

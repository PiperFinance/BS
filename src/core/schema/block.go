package schema

import "time"

type BlockTask struct {
	BlockNumber uint64 `bson:"no"`
	ChainId     int64  `bson:"chain"`
}

// MBlock Block Stored At Mongo
type BlockM struct {
	ScannerStatus `bson:"status"`
	BlockNumber   uint64    `bson:"no"`
	UpdatedAt     time.Time `bson:"c_at"`
	StartedAt     time.Time `bson:"s_at"`
	ChainId       int64     `bson:"chain"`
}

func (bm *BlockM) SetScanned() BlockM {
	return BlockM{
		ScannerStatus: Scanned,
		UpdatedAt:     time.Now(),
		StartedAt:     time.Now(),
	}
}

func (bm *BlockM) SetFetched() BlockM {
	return BlockM{
		ScannerStatus: Fetched,
		UpdatedAt:     time.Now(),
	}
}

func (bm *BlockM) SetParsed() BlockM {
	return BlockM{
		ScannerStatus: Scanned,
		UpdatedAt:     time.Now(),
	}
}

func (bm *BlockM) SetAdded() BlockM {
	return BlockM{
		ScannerStatus: Added,
		UpdatedAt:     time.Now(),
	}
}

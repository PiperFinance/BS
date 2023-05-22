package tasks

const (
	// LastScannedBlockKey TODO - Add a model in db to save fetched block numbers
	// TODO - use gocache - not redis ...
	// MultiChain ...
	LastScannedBlockKey = "block:lastScanned"

	FetchBlockEventsKey  = "block:fetch_events"
	BlockScanKey         = "block:scan"
	ParseBlockEventsKey  = "block:parse_events"
	UpdateUserBalanceKey = "user:update_bal"
	UpdateUserApproveKey = "user:update_aprv"
	VacuumLogsKey        = "chore:vacuum"
	VacuumLogsLockKey    = "chore:vacuum-lock"
	VacuumLogsHeight     = 100
)

package tasks

import "fmt"

const (
	lastScannedBlockKey  = "block:lastScanned"
	FetchBlockEventsKey  = "block:fetch_events"
	BlockScanKey         = "block:scan"
	ParseBlockEventsKey  = "block:parse_events"
	UpdateUserBalanceKey = "user:update_bal"
	UpdateUserApproveKey = "user:update_aprv"
	UpdateOnlineUsersKey = "user:online_user"
	VacuumLogsKey        = "chore:vacuum"
	VacuumLogsLockKey    = "chore:vacuum-lock"
	VacuumLogsHeight     = 100
)

func LastScannedBlockKey(chain int64) string {
	return fmt.Sprintf("[%d]:%s", chain, lastScannedBlockKey)
}

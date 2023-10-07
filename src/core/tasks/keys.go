package tasks

import "fmt"

const (
	lastScannedBlockKey  = "b:lastScanned"
	FetchBlockEventsKey  = "b:fetch_events"
	BlockScanKey         = "b:scan"
	ProccessBlockKey     = "b:process"
	ParseBlockEventsKey  = "b:parse_events"
	UpdateUserBalanceKey = "u:update_bal"
	UpdateUserApproveKey = "u:update_aprv"
	UpdateOnlineUsersKey = "u:online_user"
	VacuumLogsKey        = "c:vacuum"
	VacuumLogsLockKey    = "c:vacuum-lock"
	VacuumLogsHeight     = 2000
)

func LastScannedBlockKey(chain int64) string {
	return fmt.Sprintf("[%d]:%s", chain, lastScannedBlockKey)
}

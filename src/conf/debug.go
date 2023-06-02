package conf

import (
	"time"

	"github.com/PiperFinance/BS/src/utils"
)

var (
	CallCount      *utils.DebugCounter
	MultiCallCount *utils.DebugCounter
	NewBlockCount  *utils.DebugCounter
	NewUsersCount  *utils.DebugCounter
	ScanCallCount  *utils.DebugCounter
	FetchCallCount *utils.DebugCounter
)

func LoadDebugItems() {
	CallCount = utils.NewDebugCounter(
		Config.SupportedChains,
		100*time.Second,
		1*time.Hour,
		24*time.Hour,
	)
	NewBlockCount = utils.NewDebugCounter(
		Config.SupportedChains,
		100*time.Second,
		1*time.Hour,
		24*time.Hour,
	)
	MultiCallCount = utils.NewDebugCounter(
		Config.SupportedChains,
		100*time.Second,
		1*time.Hour,
		24*time.Hour,
	)

	ScanCallCount = utils.NewDebugCounter(
		Config.SupportedChains,
		100*time.Second,
		1*time.Hour,
		24*time.Hour,
	)
	FetchCallCount = utils.NewDebugCounter(
		Config.SupportedChains,
		100*time.Second,
		1*time.Hour,
		24*time.Hour,
	)
	NewUsersCount = utils.NewDebugCounter(
		Config.SupportedChains,
		100*time.Second,
		1*time.Hour,
		24*time.Hour,
	)
}

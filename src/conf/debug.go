package conf

import (
	"time"

	"github.com/PiperFinance/BS/src/utils"
)

var (
	CallCount     *utils.DebugCounter
	NewBlockCount *utils.DebugCounter
)

func LoadDebugItems() {
	CallCount = utils.NewDebugCounter(
		Config.SupportedChains,
		10*time.Second,
		100*time.Second,
		1*time.Hour,
		24*time.Hour,
	)
	NewBlockCount = utils.NewDebugCounter(
		Config.SupportedChains,
		10*time.Second,
		100*time.Second,
		1*time.Hour,
		24*time.Hour,
	)
}

package conf

import (
	"time"

	"github.com/PiperFinance/BS/src/utils"
)

var CallCount *utils.CallCounter

func LoadDebugItems() {
	CallCount = utils.NewCallCounter(
		Config.SupportedChains,
		1*time.Second,
		10*time.Second,
		100*time.Second,
		10*time.Second,
		1*time.Hour,
		24*time.Hour,
	)
}

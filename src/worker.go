package app

import (
	"github.com/PiperFinance/BS/src/core/conf"
)

func main() {
	conf.QueueScheduler.Run()
}
